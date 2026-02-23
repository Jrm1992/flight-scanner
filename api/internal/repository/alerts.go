package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jose/flight-scanner/internal/models"
)

// AlertRepo handles database operations for the alerts table.
type AlertRepo struct {
	db *sql.DB
}

// NewAlertRepo creates a new AlertRepo.
func NewAlertRepo(db *sql.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

// Create inserts a new alert record.
func (r *AlertRepo) Create(ctx context.Context, routeID string, alertPrice, triggeredPrice float64) (*models.Alert, error) {
	var alert models.Alert
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO alerts (route_id, alert_price, triggered_price)
		VALUES ($1, $2, $3)
		RETURNING id, route_id, alert_price, triggered_price, triggered_at, notified, notified_at
	`, routeID, alertPrice, triggeredPrice).Scan(
		&alert.ID, &alert.RouteID, &alert.AlertPrice, &alert.TriggeredPrice,
		&alert.TriggeredAt, &alert.Notified, &alert.NotifiedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert alert: %w", err)
	}
	return &alert, nil
}

// HasAlertToday checks if an alert already exists for this route today.
func (r *AlertRepo) HasAlertToday(ctx context.Context, routeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		-- NOTE: CURRENT_DATE uses the database timezone (UTC on Render)
		SELECT EXISTS(
			SELECT 1 FROM alerts
			WHERE route_id = $1 AND triggered_at::date = CURRENT_DATE
		)
	`, routeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check alert today: %w", err)
	}
	return exists, nil
}

// ListAll returns all alerts ordered by most recent first.
func (r *AlertRepo) ListAll(ctx context.Context) ([]models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, route_id, alert_price, triggered_price, triggered_at, notified, notified_at
		FROM alerts ORDER BY triggered_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanAlerts(rows)
}

// ListByRoute returns alerts for a specific route.
func (r *AlertRepo) ListByRoute(ctx context.Context, routeID string) ([]models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, route_id, alert_price, triggered_price, triggered_at, notified, notified_at
		FROM alerts WHERE route_id = $1 ORDER BY triggered_at DESC
	`, routeID)
	if err != nil {
		return nil, fmt.Errorf("list alerts by route: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanAlerts(rows)
}

// MarkRead marks an alert as notified/read.
func (r *AlertRepo) MarkRead(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE alerts SET notified = TRUE, notified_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark alert read: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanAlerts(rows *sql.Rows) ([]models.Alert, error) {
	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(
			&a.ID, &a.RouteID, &a.AlertPrice, &a.TriggeredPrice,
			&a.TriggeredAt, &a.Notified, &a.NotifiedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
