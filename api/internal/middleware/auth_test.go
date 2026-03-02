package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockValidator struct {
	userID string
	err    error
}

func (m *mockValidator) ValidateToken(_ string) (string, error) {
	return m.userID, m.err
}

func TestRequireAuth_ValidToken(t *testing.T) {
	v := &mockValidator{userID: "u-123"}
	wrap := RequireAuth(v)

	var gotUserID string
	handler := wrap(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUserID != "u-123" {
		t.Fatalf("expected user ID u-123, got %s", gotUserID)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	v := &mockValidator{userID: "u-123"}
	wrap := RequireAuth(v)

	handler := wrap(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	v := &mockValidator{err: http.ErrAbortHandler} // any error
	wrap := RequireAuth(v)

	handler := wrap(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_NoBearerPrefix(t *testing.T) {
	v := &mockValidator{userID: "u-123"}
	wrap := RequireAuth(v)

	handler := wrap(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
