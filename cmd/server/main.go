package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
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
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("starting flight-scanner (env=%s)", cfg.Env)

	// Connect to PostgreSQL
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("connected to PostgreSQL")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations complete")

	// Repositories
	routeRepo := repository.NewRouteRepo(db)
	priceHistoryRepo := repository.NewPriceHistoryRepo(db)
	alertRepo := repository.NewAlertRepo(db)

	// SerpApi (Google Flights) client
	flightClient := flightapi.NewClient(cfg.SerpAPIKey)

	// Background price monitor
	mon := monitor.New(routeRepo, priceHistoryRepo, alertRepo, flightClient)
	monCtx, monCancel := context.WithCancel(context.Background())
	defer monCancel()

	if err := mon.Start(monCtx); err != nil {
		log.Printf("warning: failed to start monitor: %v", err)
	}

	// Handlers
	routeHandler := handler.NewRouteHandler(routeRepo, mon)
	searchHandler := handler.NewSearchHandler(flightClient)
	historyHandler := handler.NewHistoryHandler(priceHistoryRepo)
	alertHandler := handler.NewAlertHandler(alertRepo)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","monitoring":%d}`, mon.RunningCount())
	})

	// /api/routes/ handles both route CRUD and history (dispatched by path)
	mux.HandleFunc("/api/routes/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/history") {
			historyHandler.ServeHTTP(w, r)
			return
		}
		routeHandler.ServeHTTP(w, r)
	})
	mux.Handle("/api/routes", routeHandler)
	mux.Handle("/api/search/flights", searchHandler)
	mux.Handle("/api/alerts/", alertHandler)
	mux.Handle("/api/alerts", alertHandler)

	// CORS middleware
	cors := middleware.CORS(cfg.FrontendURL)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	server := &http.Server{Addr: addr, Handler: cors(mux)}

	// Graceful shutdown: listen for SIGINT/SIGTERM
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received %s, shutting down...", sig)
		mon.StopAll()
		monCancel()
		server.Close()
	}()

	log.Printf("server listening on %s", addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	log.Println("server stopped")
}
