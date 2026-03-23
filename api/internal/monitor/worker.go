package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
)

// worker monitors prices for a single route in its own goroutine.
type worker struct {
	route        models.Route
	flightClient flightSearcher
	priceHistory priceHistoryStore
	alerts       alertStore

	ctx    context.Context
	cancel context.CancelFunc
}

func newWorker(parentCtx context.Context, route models.Route, fc flightSearcher, ph priceHistoryStore, al alertStore) *worker {
	ctx, cancel := context.WithCancel(parentCtx)
	return &worker{
		route:        route,
		flightClient: fc,
		priceHistory: ph,
		alerts:       al,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// run is the main loop. It ticks at the route's configured frequency, fetches prices,
// stores them, and checks alert thresholds. It exits when the context is cancelled.
func (w *worker) run() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("worker panic recovered", "route_id", w.route.ID, "origin", w.route.Origin, "destination", w.route.Destination, "panic", r)
		}
	}()

	freq := time.Duration(w.route.CheckFrequencyMinutes) * time.Minute
	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	// Do an immediate first check, then wait for ticks.
	w.check()

	for {
		select {
		case <-w.ctx.Done():
			slog.Info("worker stopped", "origin", w.route.Origin, "destination", w.route.Destination)
			return
		case <-ticker.C:
			w.check()
		}
	}
}

// stop cancels the worker's context, causing the goroutine to exit.
func (w *worker) stop() {
	w.cancel()
}

// check performs a single price fetch + store + alert evaluation cycle.
func (w *worker) check() {
	results, err := w.fetchPrices()
	if err != nil {
		slog.Error("fetch error", "origin", w.route.Origin, "destination", w.route.Destination, "err", err)
		return
	}

	if len(results) == 0 {
		slog.Info("no results", "origin", w.route.Origin, "destination", w.route.Destination)
		return
	}

	minPrice, maxPrice, avgPrice, topAirline := aggregateResults(results)

	// Persist price snapshot
	if err := w.priceHistory.Insert(w.ctx, w.route.ID, minPrice, maxPrice, avgPrice, topAirline); err != nil {
		slog.Error("price insert error", "origin", w.route.Origin, "destination", w.route.Destination, "err", err)
		return
	}

	slog.Info("price check", "origin", w.route.Origin, "destination", w.route.Destination, "min", minPrice, "max", maxPrice, "avg", avgPrice)

	// Check alert threshold
	if minPrice < w.route.AlertPrice {
		w.tryCreateAlert(minPrice)
	}
}

// parseDate parses a date string that may come as "2006-01-02" or "2006-01-02T00:00:00Z" from Postgres.
func parseDate(s string) (time.Time, error) {
	if len(s) >= 10 {
		return time.Parse("2006-01-02", s[:10])
	}
	return time.Parse("2006-01-02", s)
}

// fetchPrices searches Google Flights for this route's configured travel dates.
func (w *worker) fetchPrices() ([]flightapi.FlightResult, error) {
	departure, err := parseDate(w.route.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("parse departure_date: %w", err)
	}

	// Stop monitoring if departure date has passed
	if departure.Before(time.Now().Truncate(24 * time.Hour)) {
		slog.Info("departure date passed, stopping monitor", "route_id", w.route.ID, "departure_date", w.route.DepartureDate)
		return nil, nil
	}

	params := flightapi.SearchParams{
		DepartureID:  w.route.Origin,
		ArrivalID:    w.route.Destination,
		OutboundDate: departure,
		Currency:     w.route.Currency,
		Adults:       1,
		TravelClass:  1, // economy
	}

	if w.route.ReturnDate != nil {
		ret, err := parseDate(*w.route.ReturnDate)
		if err == nil {
			params.ReturnDate = &ret
		}
	}

	result, err := w.flightClient.Search(w.ctx, params)
	if err != nil {
		return nil, err
	}
	return result.Flights, nil
}

// tryCreateAlert creates an alert if one hasn't already been created today for this route.
func (w *worker) tryCreateAlert(triggeredPrice float64) {
	exists, err := w.alerts.HasAlertToday(w.ctx, w.route.ID)
	if err != nil {
		slog.Error("alert check error", "origin", w.route.Origin, "destination", w.route.Destination, "err", err)
		return
	}
	if exists {
		return // max 1 alert per route per day
	}

	alert, err := w.alerts.Create(w.ctx, w.route.ID, w.route.AlertPrice, triggeredPrice)
	if err != nil {
		slog.Error("alert create error", "origin", w.route.Origin, "destination", w.route.Destination, "err", err)
		return
	}

	slog.Info("price alert triggered", "origin", w.route.Origin, "destination", w.route.Destination,
		"price", triggeredPrice, "threshold", w.route.AlertPrice, "alert_id", alert.ID)
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
