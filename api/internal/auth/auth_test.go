package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/jose/flight-scanner/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) Create(_ context.Context, email, passwordHash, name string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &models.User{ID: "u-1", Email: email, PasswordHash: passwordHash, Name: name}, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func TestRegister_Success(t *testing.T) {
	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{})
	resp, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{})

	tests := []struct {
		name string
		req  models.RegisterRequest
	}{
		{"empty email", models.RegisterRequest{Email: "", Password: "password123", Name: "Test"}},
		{"invalid email", models.RegisterRequest{Email: "notanemail", Password: "password123", Name: "Test"}},
		{"short password", models.RegisterRequest{Email: "test@example.com", Password: "short", Name: "Test"}},
		{"empty name", models.RegisterRequest{Email: "test@example.com", Password: "password123", Name: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	// Generate a bcrypt hash for "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{
		user: &models.User{
			ID:           "u-1",
			Email:        "test@example.com",
			PasswordHash: string(hash),
			Name:         "Test User",
		},
	})
	resp, err := svc.Login(context.Background(), models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{
		err: errors.New("not found"),
	})
	_, err := svc.Login(context.Background(), models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateToken(t *testing.T) {
	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{})

	token, err := svc.generateToken("u-123")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	userID, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if userID != "u-123" {
		t.Fatalf("expected u-123, got %s", userID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := NewService("test-secret-32-chars-long-enough!", &mockUserRepo{})

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewService("secret-one-32-chars-long-enough!", &mockUserRepo{})
	svc2 := NewService("secret-two-32-chars-long-enough!", &mockUserRepo{})

	token, _ := svc1.generateToken("u-123")
	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
