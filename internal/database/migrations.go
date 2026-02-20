package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// migrations is an ordered list of SQL statements that set up the schema.
// Each migration runs inside a transaction. New migrations should be appended
// at the end — never modify or reorder existing entries.
var migrations = []struct {
	name string
	sql  string
}{
	{
		name: "create_routes_table",
		sql: `CREATE TABLE IF NOT EXISTS routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			origin VARCHAR(3) NOT NULL,
			destination VARCHAR(3) NOT NULL,
			alert_price DECIMAL(10, 2) NOT NULL,
			check_frequency_minutes INT DEFAULT 60,
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
	},
	{
		name: "create_price_history_table",
		sql: `CREATE TABLE IF NOT EXISTS price_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
			min_price DECIMAL(10, 2) NOT NULL,
			max_price DECIMAL(10, 2) NOT NULL,
			avg_price DECIMAL(10, 2) NOT NULL,
			airline VARCHAR(50),
			checked_at TIMESTAMP DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_route_checked ON price_history(route_id, checked_at);`,
	},
	{
		name: "create_alerts_table",
		sql: `CREATE TABLE IF NOT EXISTS alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
			alert_price DECIMAL(10, 2) NOT NULL,
			triggered_price DECIMAL(10, 2) NOT NULL,
			triggered_at TIMESTAMP DEFAULT NOW(),
			notified BOOLEAN DEFAULT FALSE,
			notified_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_route_alerts ON alerts(route_id, triggered_at);`,
	},
}

// RunMigrations executes all pending migrations in order.
// It uses a simple migrations tracking table to avoid re-running completed migrations.
func RunMigrations(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create migrations tracking table
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT NOW()
	);`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		// Check if already applied
		var exists bool
		err := db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)", m.name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if exists {
			continue
		}

		// Run migration in a transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", m.name, err)
		}

		if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %s: %w", m.name, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", m.name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.name, err)
		}

		log.Printf("[migration] applied: %s", m.name)
	}

	return nil
}
