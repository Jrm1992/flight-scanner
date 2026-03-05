package repository

import (
	"context"
	"database/sql"
	"errors"
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

const routeColumns = `id, user_id, origin, destination, departure_date, return_date, alert_price, check_frequency_minutes, status, created_at, updated_at`

func scanRoute(scanner interface{ Scan(...any) error }) (*models.Route, error) {
	var route models.Route
	err := scanner.Scan(
		&route.ID, &route.UserID, &route.Origin, &route.Destination,
		&route.DepartureDate, &route.ReturnDate,
		&route.AlertPrice, &route.CheckFrequencyMinutes, &route.Status,
		&route.CreatedAt, &route.UpdatedAt,
	)
	return &route, err
}

// Create inserts a new route and returns it with the generated ID.
func (r *RouteRepo) Create(ctx context.Context, userID string, req models.CreateRouteRequest) (*models.Route, error) {
	freq := req.CheckFrequencyMinutes
	if freq <= 0 {
		freq = 60
	}

	route, err := scanRoute(r.db.QueryRowContext(ctx, `
		INSERT INTO routes (user_id, origin, destination, departure_date, return_date, alert_price, check_frequency_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+routeColumns,
		userID, req.Origin, req.Destination, req.DepartureDate, req.ReturnDate, req.AlertPrice, freq,
	))
	if err != nil {
		return nil, fmt.Errorf("insert route: %w", err)
	}
	return route, nil
}

// GetByID retrieves a single route by ID, scoped to the user.
func (r *RouteRepo) GetByID(ctx context.Context, userID, id string) (*models.Route, error) {
	route, err := scanRoute(r.db.QueryRowContext(ctx,
		`SELECT `+routeColumns+` FROM routes WHERE id = $1 AND user_id = $2`, id, userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get route: %w", err)
	}
	return route, nil
}

// ListActive returns all routes with status 'active' across all users (used by monitor).
func (r *RouteRepo) ListActive(ctx context.Context) ([]models.Route, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+routeColumns+` FROM routes WHERE status = 'active' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list active routes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanRoutes(rows)
}

// ListAll returns all routes for a specific user.
func (r *RouteRepo) ListAll(ctx context.Context, userID string) ([]models.Route, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+routeColumns+` FROM routes WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanRoutes(rows)
}

// Update modifies a route's alert_price and/or check_frequency_minutes, scoped to user.
func (r *RouteRepo) Update(ctx context.Context, userID, id string, req models.UpdateRouteRequest) (*models.Route, error) {
	route, err := scanRoute(r.db.QueryRowContext(ctx, `
		UPDATE routes SET
			alert_price = COALESCE($3, alert_price),
			check_frequency_minutes = COALESCE($4, check_frequency_minutes),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING `+routeColumns,
		id, userID, req.AlertPrice, req.CheckFrequencyMinutes,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update route: %w", err)
	}
	return route, nil
}

// SetStatus changes a route's status (e.g. "active", "paused"), scoped to user.
func (r *RouteRepo) SetStatus(ctx context.Context, userID, id, status string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE routes SET status = $3, updated_at = NOW() WHERE id = $1 AND user_id = $2
	`, id, userID, status)
	if err != nil {
		return fmt.Errorf("set route status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a route by ID, scoped to user. Price history is cascade-deleted.
func (r *RouteRepo) Delete(ctx context.Context, userID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM routes WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete route: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanRoutes(rows *sql.Rows) ([]models.Route, error) {
	var routes []models.Route
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}
		routes = append(routes, *route)
	}
	return routes, rows.Err()
}
