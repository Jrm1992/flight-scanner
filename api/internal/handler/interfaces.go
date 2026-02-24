package handler

import (
	"context"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/monitor"
	"github.com/jose/flight-scanner/internal/repository"
)

// Compile-time interface checks.
var (
	_ RouteRepository        = (*repository.RouteRepo)(nil)
	_ AlertRepository        = (*repository.AlertRepo)(nil)
	_ PriceHistoryRepository = (*repository.PriceHistoryRepo)(nil)
	_ FlightSearcher         = (*flightapi.Client)(nil)
	_ RouteMonitor           = (*monitor.Monitor)(nil)
)

// RouteRepository defines the methods handlers need from the route store.
type RouteRepository interface {
	Create(ctx context.Context, req models.CreateRouteRequest) (*models.Route, error)
	ListAll(ctx context.Context) ([]models.Route, error)
	Update(ctx context.Context, id string, req models.UpdateRouteRequest) (*models.Route, error)
	Delete(ctx context.Context, id string) error
	SetStatus(ctx context.Context, id, status string) error
	GetByID(ctx context.Context, id string) (*models.Route, error)
}

// AlertRepository defines the methods handlers need from the alert store.
type AlertRepository interface {
	ListAll(ctx context.Context) ([]models.Alert, error)
	ListByRoute(ctx context.Context, routeID string) ([]models.Alert, error)
	MarkRead(ctx context.Context, id string) error
}

// PriceHistoryRepository defines the methods handlers need from the price history store.
type PriceHistoryRepository interface {
	GetByRoute(ctx context.Context, routeID string, days int) ([]models.PriceHistory, error)
	GetStats(ctx context.Context, routeID string, days int) (*models.PriceStats, error)
	GetLatestPrices(ctx context.Context, routeIDs []string) (map[string]models.PriceHistory, error)
}

// FlightSearcher defines the method handlers need for flight searches.
type FlightSearcher interface {
	Search(ctx context.Context, params flightapi.SearchParams) ([]flightapi.FlightResult, error)
}

// RouteMonitor defines the methods handlers need for managing route monitoring.
type RouteMonitor interface {
	StartRoute(route models.Route)
	StopRoute(routeID string)
	RestartRoute(route models.Route)
}
