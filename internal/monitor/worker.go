package monitor

import (
	"context"
	"log"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

// worker monitors prices for a single route in its own goroutine.
type worker struct {
	route        models.Route
	flightClient *flightapi.Client
	priceHistory *repository.PriceHistoryRepo
	alerts       *repository.AlertRepo

	cancel context.CancelFunc
}

func newWorker(route models.Route, fc *flightapi.Client, ph *repository.PriceHistoryRepo, al *repository.AlertRepo) *worker {
	return &worker{
		route:        route,
		flightClient: fc,
		priceHistory: ph,
		alerts:       al,
	}
}

// run is the main loop. It ticks at the route's configured frequency, fetches prices,
// stores them, and checks alert thresholds. It exits when the context is cancelled.
func (w *worker) run(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	w.cancel = cancel

	freq := time.Duration(w.route.CheckFrequencyMinutes) * time.Minute
	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	// Do an immediate first check, then wait for ticks.
	w.check(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[worker] %s→%s stopped", w.route.Origin, w.route.Destination)
			return
		case <-ticker.C:
			w.check(ctx)
		}
	}
}

// stop cancels the worker's context, causing the goroutine to exit.
func (w *worker) stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// check performs a single price fetch + store + alert evaluation cycle.
func (w *worker) check(ctx context.Context) {
	results, err := w.fetchPrices(ctx)
	if err != nil {
		log.Printf("[worker] %s→%s fetch error: %v", w.route.Origin, w.route.Destination, err)
		return
	}

	if len(results) == 0 {
		log.Printf("[worker] %s→%s no results", w.route.Origin, w.route.Destination)
		return
	}

	minPrice, maxPrice, avgPrice, topAirline := aggregateResults(results)

	// Persist price snapshot
	if err := w.priceHistory.Insert(ctx, w.route.ID, minPrice, maxPrice, avgPrice, topAirline); err != nil {
		log.Printf("[worker] %s→%s price insert error: %v", w.route.Origin, w.route.Destination, err)
		return
	}

	log.Printf("[worker] %s→%s price: min=%.2f max=%.2f avg=%.2f",
		w.route.Origin, w.route.Destination, minPrice, maxPrice, avgPrice)

	// Check alert threshold
	if minPrice < w.route.AlertPrice {
		w.tryCreateAlert(ctx, minPrice)
	}
}

// fetchPrices searches Google Flights for this route over the next 30 days.
func (w *worker) fetchPrices(ctx context.Context) ([]flightapi.FlightResult, error) {
	now := time.Now()
	params := flightapi.SearchParams{
		DepartureID:  w.route.Origin,
		ArrivalID:    w.route.Destination,
		OutboundDate: now.AddDate(0, 0, 1), // tomorrow
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1, // economy
	}
	return w.flightClient.Search(ctx, params)
}

// tryCreateAlert creates an alert if one hasn't already been created today for this route.
func (w *worker) tryCreateAlert(ctx context.Context, triggeredPrice float64) {
	exists, err := w.alerts.HasAlertToday(ctx, w.route.ID)
	if err != nil {
		log.Printf("[worker] %s→%s alert check error: %v", w.route.Origin, w.route.Destination, err)
		return
	}
	if exists {
		return // max 1 alert per route per day
	}

	alert, err := w.alerts.Create(ctx, w.route.ID, w.route.AlertPrice, triggeredPrice)
	if err != nil {
		log.Printf("[worker] %s→%s alert create error: %v", w.route.Origin, w.route.Destination, err)
		return
	}

	log.Printf("[worker] ALERT %s→%s: price %.2f < threshold %.2f (alert_id=%s)",
		w.route.Origin, w.route.Destination, triggeredPrice, w.route.AlertPrice, alert.ID)
}

// aggregateResults computes min, max, avg prices and the cheapest airline from results.
func aggregateResults(results []flightapi.FlightResult) (min, max, avg float64, airline string) {
	min = results[0].Price
	max = results[0].Price
	airline = results[0].Airline
	var total float64

	for _, r := range results {
		total += r.Price
		if r.Price < min {
			min = r.Price
			airline = r.Airline
		}
		if r.Price > max {
			max = r.Price
		}
	}

	avg = total / float64(len(results))
	return
}
