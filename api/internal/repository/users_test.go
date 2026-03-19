package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

var userCols = []string{"id", "email", "password_hash", "name", "created_at"}

func TestUserRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("test@example.com", "hashed-pw", "Test User").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"user-1", "test@example.com", "hashed-pw", "Test User", now,
		))

	user, err := repo.Create(context.Background(), "test@example.com", "hashed-pw", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected ID user-1, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", user.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepo(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("dup@example.com", "hashed-pw", "Dup User").
		WillReturnError(&pq.Error{Code: "23505"})

	_, err = repo.Create(context.Background(), "dup@example.com", "hashed-pw", "Dup User")
	if err != ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email = .+`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"user-1", "test@example.com", "hashed-pw", "Test User", now,
		))

	user, err := repo.GetByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected ID user-1, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepo(db)

	mock.ExpectQuery(`SELECT .+ FROM users WHERE email = .+`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByEmail(context.Background(), "missing@example.com")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM users WHERE id = .+`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"user-1", "test@example.com", "hashed-pw", "Test User", now,
		))

	user, err := repo.GetByID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected ID user-1, got %s", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", user.Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
