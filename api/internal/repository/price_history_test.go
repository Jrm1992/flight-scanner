package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var priceCols = []string{
	"id", "route_id", "min_price", "max_price", "avg_price", "airline", "checked_at",
}

func TestPriceHistoryRepo_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)

	mock.ExpectExec(`INSERT INTO price_history`).
		WithArgs("route-1", 100.0, 300.0, 200.0, "Delta").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Insert(context.Background(), "route-1", 100.0, 300.0, 200.0, "Delta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPriceHistoryRepo_GetByRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT ph.id, ph.route_id, ph.min_price, ph.max_price, ph.avg_price`).
		WithArgs("route-1", "user-1", 30).
		WillReturnRows(sqlmock.NewRows(priceCols).
			AddRow("ph-1", "route-1", 100.0, 300.0, 200.0, "Delta", now).
			AddRow("ph-2", "route-1", 110.0, 280.0, 190.0, "United", now),
		)

	history, err := repo.GetByRoute(context.Background(), "user-1", "route-1", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 records, got %d", len(history))
	}
	if history[0].MinPrice != 100.0 {
		t.Errorf("expected min price 100, got %f", history[0].MinPrice)
	}
	if history[1].Airline != "United" {
		t.Errorf("expected airline United, got %s", history[1].Airline)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPriceHistoryRepo_GetLatestPrice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, route_id, min_price, max_price, avg_price`).
		WithArgs("route-1").
		WillReturnRows(sqlmock.NewRows(priceCols).AddRow(
			"ph-1", "route-1", 99.0, 250.0, 175.0, "Delta", now,
		))

	ph, err := repo.GetLatestPrice(context.Background(), "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ph.MinPrice != 99.0 {
		t.Errorf("expected min price 99, got %f", ph.MinPrice)
	}
	if ph.RouteID != "route-1" {
		t.Errorf("expected route_id route-1, got %s", ph.RouteID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPriceHistoryRepo_GetLatestPrice_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)

	mock.ExpectQuery(`SELECT id, route_id, min_price, max_price, avg_price`).
		WithArgs("route-1").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetLatestPrice(context.Background(), "route-1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPriceHistoryRepo_GetLatestPrices(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT DISTINCT ON`).
		WithArgs("user-1", "route-1", "route-2").
		WillReturnRows(sqlmock.NewRows(priceCols).
			AddRow("ph-1", "route-1", 100.0, 300.0, 200.0, "Delta", now).
			AddRow("ph-2", "route-2", 150.0, 400.0, 275.0, "United", now),
		)

	result, err := repo.GetLatestPrices(context.Background(), "user-1", []string{"route-1", "route-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if ph, ok := result["route-1"]; !ok {
		t.Error("expected route-1 in result")
	} else if ph.MinPrice != 100.0 {
		t.Errorf("expected min price 100 for route-1, got %f", ph.MinPrice)
	}
	if ph, ok := result["route-2"]; !ok {
		t.Error("expected route-2 in result")
	} else if ph.Airline != "United" {
		t.Errorf("expected airline United for route-2, got %s", ph.Airline)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPriceHistoryRepo_GetLatestPrices_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)

	result, err := repo.GetLatestPrices(context.Background(), "user-1", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestPriceHistoryRepo_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPriceHistoryRepo(db)

	mock.ExpectQuery(`SELECT COALESCE\(MIN\(ph.min_price\), 0\), COALESCE\(MAX\(ph.max_price\), 0\), COALESCE\(AVG\(ph.avg_price\), 0\)`).
		WithArgs("route-1", "user-1", 30).
		WillReturnRows(sqlmock.NewRows([]string{"min", "max", "avg"}).AddRow(80.0, 350.0, 210.0))

	stats, err := repo.GetStats(context.Background(), "user-1", "route-1", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.MinPrice != 80.0 {
		t.Errorf("expected min price 80, got %f", stats.MinPrice)
	}
	if stats.MaxPrice != 350.0 {
		t.Errorf("expected max price 350, got %f", stats.MaxPrice)
	}
	if stats.AvgPrice != 210.0 {
		t.Errorf("expected avg price 210, got %f", stats.AvgPrice)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
