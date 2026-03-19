package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var alertCols = []string{
	"id", "route_id", "alert_price", "triggered_price", "triggered_at", "notified", "notified_at",
}

func TestAlertRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO alerts`).
		WithArgs("route-1", 200.0, 180.0).
		WillReturnRows(sqlmock.NewRows(alertCols).AddRow(
			"alert-1", "route-1", 200.0, 180.0, now, false, nil,
		))

	alert, err := repo.Create(context.Background(), "route-1", 200.0, 180.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert.ID != "alert-1" {
		t.Errorf("expected ID alert-1, got %s", alert.ID)
	}
	if alert.TriggeredPrice != 180.0 {
		t.Errorf("expected triggered price 180, got %f", alert.TriggeredPrice)
	}
	if alert.Notified {
		t.Error("expected notified to be false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAlertRepo_HasAlertToday_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("route-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.HasAlertToday(context.Background(), "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAlertRepo_HasAlertToday_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("route-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.HasAlertToday(context.Background(), "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists to be false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAlertRepo_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT a.id, a.route_id, a.alert_price, a.triggered_price, a.triggered_at, a.notified, a.notified_at`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(alertCols).
			AddRow("a-1", "route-1", 200.0, 180.0, now, false, nil).
			AddRow("a-2", "route-2", 300.0, 250.0, now, true, &now),
		)

	alerts, err := repo.ListAll(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}
	if alerts[0].ID != "a-1" {
		t.Errorf("expected first alert a-1, got %s", alerts[0].ID)
	}
	if alerts[1].Notified != true {
		t.Error("expected second alert to be notified")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAlertRepo_ListByRoute(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT a.id, a.route_id, a.alert_price, a.triggered_price, a.triggered_at, a.notified, a.notified_at`).
		WithArgs("route-1", "user-1").
		WillReturnRows(sqlmock.NewRows(alertCols).
			AddRow("a-1", "route-1", 200.0, 180.0, now, false, nil),
		)

	alerts, err := repo.ListByRoute(context.Background(), "user-1", "route-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RouteID != "route-1" {
		t.Errorf("expected route_id route-1, got %s", alerts[0].RouteID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAlertRepo_MarkRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAlertRepo(db)

	mock.ExpectExec(`UPDATE alerts SET notified = TRUE`).
		WithArgs("alert-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.MarkRead(context.Background(), "user-1", "alert-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
