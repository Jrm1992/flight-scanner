package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jose/flight-scanner/internal/auth"
	"github.com/jose/flight-scanner/internal/config"
	"github.com/jose/flight-scanner/internal/database"
	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/handler"
	"github.com/jose/flight-scanner/internal/middleware"
	"github.com/jose/flight-scanner/internal/monitor"
	"github.com/jose/flight-scanner/internal/repository"
	"github.com/jose/flight-scanner/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Load .env file in development (silently ignore if missing)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	slog.Info("starting flight-scanner", "env", cfg.Env)

	// Initialize OpenTelemetry
	if err := telemetry.Init(context.Background(), "flight-scanner-api", cfg.OTLPEndpoint); err != nil {
		slog.Warn("failed to initialize telemetry", "err", err)
	}
	defer telemetry.Shutdown(context.Background())

	// Connect to PostgreSQL
	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("db close error", "err", err)
		}
	}()
	slog.Info("connected to PostgreSQL")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1) //nolint:gocritic // defers are for cleanup; migration failure is fatal
	}
	slog.Info("migrations complete")

	// Repositories
	routeRepo := repository.NewRouteRepo(db)
	priceHistoryRepo := repository.NewPriceHistoryRepo(db)
	alertRepo := repository.NewAlertRepo(db)
	userRepo := repository.NewUserRepo(db)

	// Auth service
	authService := auth.NewService(cfg.JWTSecret, userRepo)

	// SerpApi (Google Flights) client
	flightClient := flightapi.NewClient(cfg.SerpAPIKey)

	// Background price monitor
	mon := monitor.New(routeRepo, priceHistoryRepo, alertRepo, flightClient)
	monCtx, monCancel := context.WithCancel(context.Background())
	defer monCancel()

	if err := mon.Start(monCtx); err != nil {
		slog.Warn("failed to start monitor", "err", err)
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	routeHandler := handler.NewRouteHandler(routeRepo, mon, priceHistoryRepo)
	searchHandler := handler.NewSearchHandler(flightClient)
	historyHandler := handler.NewHistoryHandler(priceHistoryRepo)
	alertHandler := handler.NewAlertHandler(alertRepo)

	// Routes — Go 1.22+ pattern matching
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"status":"ok","monitoring":%d}`, mon.RunningCount()); err != nil {
			slog.Error("health write error", "err", err)
		}
	})

	// Public routes
	authHandler.RegisterRoutes(mux)

	// Protected routes — wrapped with auth middleware
	auth := middleware.RequireAuth(authService)
	mux.HandleFunc("GET /api/routes", auth(routeHandler.List))
	mux.HandleFunc("POST /api/routes", auth(routeHandler.Create))
	mux.HandleFunc("PUT /api/routes/{id}", auth(routeHandler.Update))
	mux.HandleFunc("DELETE /api/routes/{id}", auth(routeHandler.Delete))
	mux.HandleFunc("PATCH /api/routes/{id}/pause", auth(routeHandler.Pause))
	mux.HandleFunc("PATCH /api/routes/{id}/resume", auth(routeHandler.Resume))
	mux.HandleFunc("GET /api/routes/{id}/history", auth(historyHandler.GetHistory))
	mux.HandleFunc("GET /api/routes/{id}/history/export", auth(historyHandler.Export))
	mux.HandleFunc("POST /api/search/flights", auth(searchHandler.Search))
	mux.HandleFunc("GET /api/search/airports", auth(searchHandler.Autocomplete))
	mux.HandleFunc("GET /api/alerts", auth(alertHandler.List))
	mux.HandleFunc("PATCH /api/alerts/{id}/mark-read", auth(alertHandler.MarkRead))

	// Middleware: CORS + OpenTelemetry HTTP instrumentation
	cors := middleware.CORS(cfg.FrontendURL)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      otelhttp.NewHandler(cors(mux), "flight-scanner-api"),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: listen for SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "addr", addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)

	mon.StopAll()
	monCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "err", err)
	}

	slog.Info("server stopped")
}
