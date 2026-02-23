package monitor

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jose/flight-scanner/internal/models"
)

// Monitor manages background goroutines that track flight prices for active routes.
type Monitor struct {
	routes       routeStore
	priceHistory priceHistoryStore
	alerts       alertStore
	flightClient flightSearcher

	ctx     context.Context // long-lived context for all workers
	mu      sync.RWMutex
	workers map[string]*worker // route ID -> worker
}

// New creates a Monitor with the given dependencies.
func New(routes routeStore, priceHistory priceHistoryStore, alerts alertStore, flightClient flightSearcher) *Monitor {
	return &Monitor{
		routes:       routes,
		priceHistory: priceHistory,
		alerts:       alerts,
		flightClient: flightClient,
		workers:      make(map[string]*worker),
	}
}

// Start loads all active routes and starts a monitoring goroutine for each.
// The provided context is stored and used as the parent for all worker goroutines.
func (m *Monitor) Start(ctx context.Context) error {
	m.ctx = ctx

	active, err := m.routes.ListActive(ctx)
	if err != nil {
		return err
	}

	slog.Info("starting active routes", "count", len(active))
	for _, route := range active {
		m.StartRoute(route)
	}
	return nil
}

// StartRoute begins monitoring a single route. Safe to call if already running (no-op).
func (m *Monitor) StartRoute(route models.Route) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workers[route.ID]; exists {
		return
	}

	w := newWorker(m.ctx, route, m.flightClient, m.priceHistory, m.alerts)
	m.workers[route.ID] = w
	go w.run()

	slog.Info("started worker", "origin", route.Origin, "destination", route.Destination, "id", route.ID, "freq_min", route.CheckFrequencyMinutes)
}

// StopRoute cancels the monitoring goroutine for a given route.
func (m *Monitor) StopRoute(routeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, exists := m.workers[routeID]; exists {
		w.stop()
		delete(m.workers, routeID)
		slog.Info("stopped worker", "route_id", routeID)
	}
}

// StopAll cancels all monitoring goroutines. Called during graceful shutdown.
func (m *Monitor) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, w := range m.workers {
		w.stop()
		delete(m.workers, id)
	}
	slog.Info("all workers stopped")
}

// RunningCount returns the number of currently active workers.
func (m *Monitor) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.workers)
}

// IsRunning checks if a specific route is being monitored.
func (m *Monitor) IsRunning(routeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.workers[routeID]
	return exists
}

// RestartRoute stops and restarts monitoring for a route (e.g. after config update).
func (m *Monitor) RestartRoute(route models.Route) {
	m.mu.Lock()
	if w, exists := m.workers[route.ID]; exists {
		w.stop()
		delete(m.workers, route.ID)
	}
	m.mu.Unlock()
	m.StartRoute(route)
}
