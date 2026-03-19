package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jose/flight-scanner/internal/models"
)

var routeCols = []string{
	"id", "user_id", "origin", "destination", "departure_date", "return_date",
	"alert_price", "check_frequency_minutes", "status", "created_at", "updated_at",
}

func TestRouteRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)
	now := time.Now()
	returnDate := "2026-04-20"

	mock.ExpectQuery(`INSERT INTO routes`).
		WithArgs("user-1", "JFK", "LAX", "2026-04-10", &returnDate, 150.0, 60).
		WillReturnRows(sqlmock.NewRows(routeCols).AddRow(
			"route-1", "user-1", "JFK", "LAX", "2026-04-10", &returnDate,
			150.0, 60, "active", now, now,
		))

	req := models.CreateRouteRequest{
		Origin:                "JFK",
		Destination:           "LAX",
		DepartureDate:         "2026-04-10",
		ReturnDate:            &returnDate,
		AlertPrice:            150.0,
		CheckFrequencyMinutes: 60,
	}
	route, err := repo.Create(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.ID != "route-1" {
		t.Errorf("expected ID route-1, got %s", route.ID)
	}
	if route.Origin != "JFK" {
		t.Errorf("expected origin JFK, got %s", route.Origin)
	}
	if route.Status != "active" {
		t.Errorf("expected status active, got %s", route.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM routes WHERE id = .+ AND user_id = .+`).
		WithArgs("route-1", "user-1").
		WillReturnRows(sqlmock.NewRows(routeCols).AddRow(
			"route-1", "user-1", "JFK", "LAX", "2026-04-10", nil,
			200.0, 30, "active", now, now,
		))

	route, err := repo.GetByID(context.Background(), "user-1", "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.ID != "route-1" {
		t.Errorf("expected ID route-1, got %s", route.ID)
	}
	if route.AlertPrice != 200.0 {
		t.Errorf("expected alert price 200, got %f", route.AlertPrice)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)

	mock.ExpectQuery(`SELECT .+ FROM routes WHERE id = .+ AND user_id = .+`).
		WithArgs("no-route", "user-1").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), "user-1", "no-route")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM routes WHERE user_id = .+ ORDER BY created_at DESC`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(routeCols).
			AddRow("r-1", "user-1", "JFK", "LAX", "2026-04-10", nil, 100.0, 60, "active", now, now).
			AddRow("r-2", "user-1", "SFO", "ORD", "2026-05-01", nil, 200.0, 30, "paused", now, now),
		)

	routes, err := repo.ListAll(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
	if routes[0].ID != "r-1" {
		t.Errorf("expected first route r-1, got %s", routes[0].ID)
	}
	if routes[1].Origin != "SFO" {
		t.Errorf("expected second route origin SFO, got %s", routes[1].Origin)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_ListActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM routes WHERE status = 'active' ORDER BY created_at DESC`).
		WillReturnRows(sqlmock.NewRows(routeCols).
			AddRow("r-1", "user-1", "JFK", "LAX", "2026-04-10", nil, 100.0, 60, "active", now, now),
		)

	routes, err := repo.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Status != "active" {
		t.Errorf("expected status active, got %s", routes[0].Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)
	now := time.Now()
	newPrice := 120.0
	newFreq := 15

	mock.ExpectQuery(`UPDATE routes SET`).
		WithArgs("route-1", "user-1", &newPrice, &newFreq).
		WillReturnRows(sqlmock.NewRows(routeCols).AddRow(
			"route-1", "user-1", "JFK", "LAX", "2026-04-10", nil,
			120.0, 15, "active", now, now,
		))

	req := models.UpdateRouteRequest{
		AlertPrice:            &newPrice,
		CheckFrequencyMinutes: &newFreq,
	}
	route, err := repo.Update(context.Background(), "user-1", "route-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route.AlertPrice != 120.0 {
		t.Errorf("expected alert price 120, got %f", route.AlertPrice)
	}
	if route.CheckFrequencyMinutes != 15 {
		t.Errorf("expected frequency 15, got %d", route.CheckFrequencyMinutes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)

	mock.ExpectExec(`DELETE FROM routes WHERE id = .+ AND user_id = .+`).
		WithArgs("route-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), "user-1", "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRouteRepo_SetStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRouteRepo(db)

	mock.ExpectExec(`UPDATE routes SET status = .+`).
		WithArgs("route-1", "user-1", "paused").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.SetStatus(context.Background(), "user-1", "route-1", "paused")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
