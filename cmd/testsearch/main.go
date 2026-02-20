package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/jose/flight-scanner/internal/config"
	"github.com/jose/flight-scanner/internal/database"
	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/monitor"
	"github.com/jose/flight-scanner/internal/repository"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	// --- 1. Test SerpApi flight search ---
	fmt.Println("============================================================")
	fmt.Println("  Flight Search: Rio de Janeiro (GIG) → Santiago (SCL)")
	fmt.Println("============================================================")
	fmt.Println()

	flightClient := flightapi.NewClient(cfg.SerpAPIKey)

	tomorrow := time.Now().AddDate(0, 0, 1)
	params := flightapi.SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: tomorrow,
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1, // economy
	}

	fmt.Printf("Searching flights %s → %s (date: %s)...\n\n",
		params.DepartureID, params.ArrivalID,
		params.OutboundDate.Format("2006-01-02"),
	)

	results, err := flightClient.Search(ctx, params)
	if err != nil {
		fmt.Printf("SerpApi error: %v\n", err)
		fmt.Println("(Make sure SERPAPI_KEY is set to a valid key in .env)")
		fmt.Println()
		fmt.Println("Skipping API search, proceeding to database test...")
		fmt.Println()
	} else if len(results) == 0 {
		fmt.Println("No flights found.")
		fmt.Println()
	} else {
		fmt.Printf("Found %d flight(s):\n\n", len(results))
		for i, r := range results {
			dep := "N/A"
			if !r.Departure.IsZero() {
				dep = r.Departure.Format("02 Jan 2006 15:04")
			}
			stops := "direct"
			if r.Stops == 1 {
				stops = "1 stop"
			} else if r.Stops > 1 {
				stops = fmt.Sprintf("%d stops", r.Stops)
			}
			dur := fmt.Sprintf("%dh%02dm", r.Duration/60, r.Duration%60)
			fmt.Printf("  %d. $%.0f  %-15s  %s → %s  %s  %s  %s\n",
				i+1, r.Price, r.Airline, r.DepartureCode, r.ArrivalCode, dep, dur, stops)
		}
		fmt.Println()
	}

	// --- 2. Test database: create route ---
	fmt.Println("------------------------------------------------------------")
	fmt.Println("  Database Test: Creating monitoring route GIG → SCL")
	fmt.Println("------------------------------------------------------------")
	fmt.Println()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	fmt.Println("Connected to PostgreSQL")

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	routeRepo := repository.NewRouteRepo(db)
	priceHistoryRepo := repository.NewPriceHistoryRepo(db)
	alertRepo := repository.NewAlertRepo(db)

	route, err := routeRepo.Create(ctx, models.CreateRouteRequest{
		Origin:                "GIG",
		Destination:           "SCL",
		AlertPrice:            250.00,
		CheckFrequencyMinutes: 60,
	})
	if err != nil {
		log.Fatalf("create route: %v", err)
	}

	fmt.Printf("Route created:\n")
	fmt.Printf("  ID:        %s\n", route.ID)
	fmt.Printf("  Route:     %s → %s\n", route.Origin, route.Destination)
	fmt.Printf("  Alert at:  $%.2f\n", route.AlertPrice)
	fmt.Printf("  Frequency: every %d min\n", route.CheckFrequencyMinutes)
	fmt.Printf("  Status:    %s\n", route.Status)
	fmt.Println()

	// --- 3. Test monitor: run one check cycle ---
	fmt.Println("------------------------------------------------------------")
	fmt.Println("  Monitor Test: Running one price check cycle")
	fmt.Println("------------------------------------------------------------")
	fmt.Println()

	mon := monitor.New(routeRepo, priceHistoryRepo, alertRepo, flightClient)
	mon.StartRoute(ctx, *route)
	fmt.Printf("Monitor started (workers running: %d)\n", mon.RunningCount())

	// Wait for the first check to complete
	fmt.Println("Waiting 20s for first price check...")
	time.Sleep(20 * time.Second)

	// Check if price was recorded
	latest, err := priceHistoryRepo.GetLatestPrice(ctx, route.ID)
	if err != nil {
		fmt.Printf("Error getting latest price: %v\n", err)
	} else if latest == nil {
		fmt.Println("No price recorded yet (API may have failed)")
	} else {
		fmt.Printf("\nLatest price recorded:\n")
		fmt.Printf("  Min: $%.2f\n", latest.MinPrice)
		fmt.Printf("  Max: $%.2f\n", latest.MaxPrice)
		fmt.Printf("  Avg: $%.2f\n", latest.AvgPrice)
		fmt.Printf("  Airline: %s\n", latest.Airline)
		fmt.Printf("  Checked: %s\n", latest.CheckedAt.Format(time.RFC3339))
	}

	// Check for alerts
	alerts, err := alertRepo.ListByRoute(ctx, route.ID)
	if err != nil {
		fmt.Printf("Error getting alerts: %v\n", err)
	} else if len(alerts) > 0 {
		fmt.Printf("\nALERT triggered! Price $%.2f < threshold $%.2f\n",
			alerts[0].TriggeredPrice, alerts[0].AlertPrice)
	} else {
		fmt.Println("\nNo alerts triggered (price is above threshold)")
	}

	// Stop monitor
	mon.StopAll()

	// List active routes
	active, err := routeRepo.ListActive(ctx)
	if err != nil {
		log.Fatalf("list routes: %v", err)
	}
	fmt.Printf("\nActive routes in database: %d\n", len(active))
	for _, r := range active {
		fmt.Printf("  - %s → %s (alert at $%.2f) [%s]\n", r.Origin, r.Destination, r.AlertPrice, r.ID)
	}

	fmt.Println()
	fmt.Println("Test complete!")
	fmt.Println()
	fmt.Println("------------------------------------------------------------")
	fmt.Println("To run the full server with continuous monitoring:")
	fmt.Println("  go run cmd/server/main.go")
	fmt.Println("------------------------------------------------------------")
}
