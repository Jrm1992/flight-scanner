package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jose/flight-scanner/internal/models"
)

// PriceHistoryRepo handles database operations for the price_history table.
type PriceHistoryRepo struct {
	db *sql.DB
}

// NewPriceHistoryRepo creates a new PriceHistoryRepo.
func NewPriceHistoryRepo(db *sql.DB) *PriceHistoryRepo {
	return &PriceHistoryRepo{db: db}
}

// Insert records a new price snapshot for a route.
func (r *PriceHistoryRepo) Insert(ctx context.Context, routeID string, minPrice, maxPrice, avgPrice float64, airline string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO price_history (route_id, min_price, max_price, avg_price, airline)
		VALUES ($1, $2, $3, $4, $5)
	`, routeID, minPrice, maxPrice, avgPrice, airline)
	if err != nil {
		return fmt.Errorf("insert price history: %w", err)
	}
	return nil
}

// GetByRoute returns price history for a route within the given number of days.
func (r *PriceHistoryRepo) GetByRoute(ctx context.Context, routeID string, days int) ([]models.PriceHistory, error) {
	if days <= 0 {
		days = 30
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, route_id, min_price, max_price, avg_price, COALESCE(airline, ''), checked_at
		FROM price_history
		WHERE route_id = $1 AND checked_at >= NOW() - ($2 || ' days')::INTERVAL
		ORDER BY checked_at ASC
	`, routeID, fmt.Sprintf("%d", days))
	if err != nil {
		return nil, fmt.Errorf("get price history: %w", err)
	}
	defer rows.Close()

	var history []models.PriceHistory
	for rows.Next() {
		var ph models.PriceHistory
		if err := rows.Scan(&ph.ID, &ph.RouteID, &ph.MinPrice, &ph.MaxPrice, &ph.AvgPrice, &ph.Airline, &ph.CheckedAt); err != nil {
			return nil, fmt.Errorf("scan price history: %w", err)
		}
		history = append(history, ph)
	}
	return history, rows.Err()
}

// GetLatestPrice returns the most recent price entry for a route.
func (r *PriceHistoryRepo) GetLatestPrice(ctx context.Context, routeID string) (*models.PriceHistory, error) {
	var ph models.PriceHistory
	err := r.db.QueryRowContext(ctx, `
		SELECT id, route_id, min_price, max_price, avg_price, COALESCE(airline, ''), checked_at
		FROM price_history
		WHERE route_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`, routeID).Scan(&ph.ID, &ph.RouteID, &ph.MinPrice, &ph.MaxPrice, &ph.AvgPrice, &ph.Airline, &ph.CheckedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest price: %w", err)
	}
	return &ph, nil
}

// GetStats returns min, max, and average prices for a route over a period.
type PriceStats struct {
	MinPrice float64   `json:"min_price"`
	MaxPrice float64   `json:"max_price"`
	AvgPrice float64   `json:"avg_price"`
	Since    time.Time `json:"since"`
}

func (r *PriceHistoryRepo) GetStats(ctx context.Context, routeID string, days int) (*PriceStats, error) {
	if days <= 0 {
		days = 30
	}

	var stats PriceStats
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MIN(min_price), 0), COALESCE(MAX(max_price), 0), COALESCE(AVG(avg_price), 0)
		FROM price_history
		WHERE route_id = $1 AND checked_at >= NOW() - ($2 || ' days')::INTERVAL
	`, routeID, fmt.Sprintf("%d", days)).Scan(&stats.MinPrice, &stats.MaxPrice, &stats.AvgPrice)
	if err != nil {
		return nil, fmt.Errorf("get price stats: %w", err)
	}
	stats.Since = time.Now().AddDate(0, 0, -days)
	return &stats, nil
}
