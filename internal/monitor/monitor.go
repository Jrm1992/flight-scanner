package monitor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

// Monitor manages background goroutines that track flight prices for active routes.
type Monitor struct {
	routes       *repository.RouteRepo
	priceHistory *repository.PriceHistoryRepo
	alerts       *repository.AlertRepo
	flightClient *flightapi.Client

	mu      sync.Mutex
	workers map[string]*worker // route ID → worker
}

// New creates a Monitor with the given dependencies.
func New(routes *repository.RouteRepo, priceHistory *repository.PriceHistoryRepo, alerts *repository.AlertRepo, flightClient *flightapi.Client) *Monitor {
	return &Monitor{
		routes:       routes,
		priceHistory: priceHistory,
		alerts:       alerts,
		flightClient: flightClient,
		workers:      make(map[string]*worker),
	}
}

// Start loads all active routes and starts a monitoring goroutine for each.
func (m *Monitor) Start(ctx context.Context) error {
	active, err := m.routes.ListActive(ctx)
	if err != nil {
		return err
	}

	log.Printf("[monitor] starting %d active route(s)", len(active))
	for _, route := range active {
		m.StartRoute(ctx, route)
	}
	return nil
}

// StartRoute begins monitoring a single route. Safe to call if already running (no-op).
func (m *Monitor) StartRoute(ctx context.Context, route models.Route) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workers[route.ID]; exists {
		return
	}

	w := newWorker(route, m.flightClient, m.priceHistory, m.alerts)
	m.workers[route.ID] = w
	go w.run(ctx)

	log.Printf("[monitor] started worker for %s→%s (id=%s, freq=%dm)",
		route.Origin, route.Destination, route.ID, route.CheckFrequencyMinutes)
}

// StopRoute cancels the monitoring goroutine for a given route.
func (m *Monitor) StopRoute(routeID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, exists := m.workers[routeID]; exists {
		w.stop()
		delete(m.workers, routeID)
		log.Printf("[monitor] stopped worker for route %s", routeID)
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
	log.Println("[monitor] all workers stopped")
}

// RunningCount returns the number of currently active workers.
func (m *Monitor) RunningCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.workers)
}

// IsRunning checks if a specific route is being monitored.
func (m *Monitor) IsRunning(routeID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.workers[routeID]
	return exists
}

// RestartRoute stops and restarts monitoring for a route (e.g. after config update).
func (m *Monitor) RestartRoute(ctx context.Context, route models.Route) {
	m.StopRoute(route.ID)
	// Small pause to ensure clean shutdown before restart.
	time.Sleep(100 * time.Millisecond)
	m.StartRoute(ctx, route)
}
