package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_NormalRequest(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS("http://localhost:3000")(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/routes", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected origin http://localhost:3000, got %s", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected Allow-Methods header to be set")
	}
}

func TestCORS_Preflight(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called on preflight")
	})

	handler := CORS("http://localhost:3000")(inner)
	req := httptest.NewRequest(http.MethodOptions, "/api/routes", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected origin http://localhost:3000, got %s", got)
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Fatalf("expected max-age 86400, got %s", got)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := CORS("https://app.example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("expected https://app.example.com, got %s", got)
	}
}
