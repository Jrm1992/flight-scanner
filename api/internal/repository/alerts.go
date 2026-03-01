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

// Create inserts a new alert record (used by monitor, no user scoping needed).
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

// HasAlertToday checks if an alert already exists for this route today (used by monitor).
func (r *AlertRepo) HasAlertToday(ctx context.Context, routeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
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

// ListAll returns all alerts for a user, ordered by most recent first.
func (r *AlertRepo) ListAll(ctx context.Context, userID string) ([]models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.route_id, a.alert_price, a.triggered_price, a.triggered_at, a.notified, a.notified_at
		FROM alerts a
		JOIN routes rt ON rt.id = a.route_id
		WHERE rt.user_id = $1
		ORDER BY a.triggered_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanAlerts(rows)
}

// ListByRoute returns alerts for a specific route, scoped to user.
func (r *AlertRepo) ListByRoute(ctx context.Context, userID, routeID string) ([]models.Alert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.route_id, a.alert_price, a.triggered_price, a.triggered_at, a.notified, a.notified_at
		FROM alerts a
		JOIN routes rt ON rt.id = a.route_id
		WHERE a.route_id = $1 AND rt.user_id = $2
		ORDER BY a.triggered_at DESC
	`, routeID, userID)
	if err != nil {
		return nil, fmt.Errorf("list alerts by route: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanAlerts(rows)
}

// MarkRead marks an alert as notified/read, scoped to user.
func (r *AlertRepo) MarkRead(ctx context.Context, userID, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE alerts SET notified = TRUE, notified_at = NOW()
		WHERE id = $1 AND route_id IN (SELECT id FROM routes WHERE user_id = $2)
	`, id, userID)
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
