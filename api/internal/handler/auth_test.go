package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jose/flight-scanner/internal/auth"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepo implements auth.UserRepository for testing.
type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) Create(_ context.Context, email, passwordHash, name string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &models.User{
		ID:           "u-1",
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func newTestAuthHandler(repo auth.UserRepository) *AuthHandler {
	svc := auth.NewService("test-secret", repo)
	return NewAuthHandler(svc)
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{})

	body := `{"email":"test@example.com","password":"password123","name":"Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.Name != "Test User" {
		t.Fatalf("expected name Test User, got %s", resp.User.Name)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{invalid`))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if body["error"] != "invalid request body" {
		t.Fatalf("expected 'invalid request body', got %q", body["error"])
	}
}

func TestRegister_ValidationError_InvalidEmail(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{})

	body := `{"email":"not-an-email","password":"password123","name":"Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{err: repository.ErrDuplicateEmail})

	body := `{"email":"test@example.com","password":"password123","name":"Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}

	var respBody map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if respBody["error"] != "email already registered" {
		t.Fatalf("expected 'email already registered', got %q", respBody["error"])
	}
}

func TestRegister_InternalError(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{err: errors.New("database down")})

	body := `{"email":"test@example.com","password":"password123","name":"Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Login tests ---

func newHashedUserRepo(t *testing.T, email, password, name string) *mockUserRepo {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return &mockUserRepo{
		user: &models.User{
			ID:           "u-1",
			Email:        email,
			PasswordHash: string(hash),
			Name:         name,
			CreatedAt:    time.Now(),
		},
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newHashedUserRepo(t, "test@example.com", "password123", "Test User")
	h := newTestAuthHandler(repo)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.Name != "Test User" {
		t.Fatalf("expected name Test User, got %s", resp.User.Name)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{bad`))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var respBody map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if respBody["error"] != "invalid request body" {
		t.Fatalf("expected 'invalid request body', got %q", respBody["error"])
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newHashedUserRepo(t, "test@example.com", "correctpassword", "Test User")
	h := newTestAuthHandler(repo)

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	var respBody map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if respBody["error"] != "invalid email or password" {
		t.Fatalf("expected 'invalid email or password', got %q", respBody["error"])
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{err: repository.ErrNotFound})

	body := `{"email":"unknown@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// TestLogin_InternalError verifies that a non-credential repo error during login
// still results in 401, because auth.Service.Login masks all GetByEmail errors
// as ErrInvalidCredentials to avoid leaking user existence information.
// The handler's 500 path is only reachable if token generation fails internally.
func TestLogin_InternalError(t *testing.T) {
	h := newTestAuthHandler(&mockUserRepo{err: errors.New("database down")})

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	// The auth service converts all GetByEmail errors to ErrInvalidCredentials,
	// so the handler returns 401 even for internal errors.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
