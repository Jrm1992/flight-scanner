package monitor

import (
	"context"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
)

type flightSearcher interface {
	Search(ctx context.Context, params flightapi.SearchParams) (flightapi.SearchResult, error)
}

type priceHistoryStore interface {
	Insert(ctx context.Context, routeID string, minPrice, maxPrice, avgPrice float64, airline string) error
}

type alertStore interface {
	HasAlertToday(ctx context.Context, routeID string) (bool, error)
	Create(ctx context.Context, routeID string, alertPrice, triggeredPrice float64) (*models.Alert, error)
}

type routeStore interface {
	ListActive(ctx context.Context) ([]models.Route, error)
}

// Compile-time interface checks.
var _ flightSearcher = (*flightapi.Client)(nil)
