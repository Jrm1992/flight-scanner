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

	// Protected routes (will be wrapped with auth middleware in Sprint 2)
	routeHandler.RegisterRoutes(mux)
	historyHandler.RegisterRoutes(mux)
	searchHandler.RegisterRoutes(mux)
	alertHandler.RegisterRoutes(mux)

	// CORS middleware
	cors := middleware.CORS(cfg.FrontendURL)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      cors(mux),
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
