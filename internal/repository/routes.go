package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jose/flight-scanner/internal/models"
)

// RouteRepo handles database operations for the routes table.
type RouteRepo struct {
	db *sql.DB
}

// NewRouteRepo creates a new RouteRepo.
func NewRouteRepo(db *sql.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

// Create inserts a new route and returns it with the generated ID.
func (r *RouteRepo) Create(ctx context.Context, req models.CreateRouteRequest) (*models.Route, error) {
	freq := req.CheckFrequencyMinutes
	if freq <= 0 {
		freq = 60
	}

	var route models.Route
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO routes (origin, destination, alert_price, check_frequency_minutes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, origin, destination, alert_price, check_frequency_minutes, status, created_at, updated_at
	`, req.Origin, req.Destination, req.AlertPrice, freq).Scan(
		&route.ID, &route.Origin, &route.Destination, &route.AlertPrice,
		&route.CheckFrequencyMinutes, &route.Status, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert route: %w", err)
	}
	return &route, nil
}

// GetByID retrieves a single route by ID.
func (r *RouteRepo) GetByID(ctx context.Context, id string) (*models.Route, error) {
	var route models.Route
	err := r.db.QueryRowContext(ctx, `
		SELECT id, origin, destination, alert_price, check_frequency_minutes, status, created_at, updated_at
		FROM routes WHERE id = $1
	`, id).Scan(
		&route.ID, &route.Origin, &route.Destination, &route.AlertPrice,
		&route.CheckFrequencyMinutes, &route.Status, &route.CreatedAt, &route.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get route: %w", err)
	}
	return &route, nil
}

// ListActive returns all routes with status 'active'.
func (r *RouteRepo) ListActive(ctx context.Context) ([]models.Route, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, origin, destination, alert_price, check_frequency_minutes, status, created_at, updated_at
		FROM routes WHERE status = 'active'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list active routes: %w", err)
	}
	defer rows.Close()

	return scanRoutes(rows)
}

// ListAll returns all routes regardless of status.
func (r *RouteRepo) ListAll(ctx context.Context) ([]models.Route, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, origin, destination, alert_price, check_frequency_minutes, status, created_at, updated_at
		FROM routes ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	defer rows.Close()

	return scanRoutes(rows)
}

// Update modifies a route's alert_price and/or check_frequency_minutes.
func (r *RouteRepo) Update(ctx context.Context, id string, req models.UpdateRouteRequest) (*models.Route, error) {
	var route models.Route
	err := r.db.QueryRowContext(ctx, `
		UPDATE routes SET
			alert_price = COALESCE($2, alert_price),
			check_frequency_minutes = COALESCE($3, check_frequency_minutes),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, origin, destination, alert_price, check_frequency_minutes, status, created_at, updated_at
	`, id, req.AlertPrice, req.CheckFrequencyMinutes).Scan(
		&route.ID, &route.Origin, &route.Destination, &route.AlertPrice,
		&route.CheckFrequencyMinutes, &route.Status, &route.CreatedAt, &route.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update route: %w", err)
	}
	return &route, nil
}

// SetStatus changes a route's status (e.g. "active", "paused").
func (r *RouteRepo) SetStatus(ctx context.Context, id, status string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE routes SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("set route status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("route %s not found", id)
	}
	return nil
}

// Delete removes a route by ID. Price history is cascade-deleted.
func (r *RouteRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM routes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete route: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("route %s not found", id)
	}
	return nil
}

func scanRoutes(rows *sql.Rows) ([]models.Route, error) {
	var routes []models.Route
	for rows.Next() {
		var route models.Route
		if err := rows.Scan(
			&route.ID, &route.Origin, &route.Destination, &route.AlertPrice,
			&route.CheckFrequencyMinutes, &route.Status, &route.CreatedAt, &route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}
