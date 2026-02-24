package config

import (
	"testing"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("SERPAPI_KEY", "key-123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://test:test@localhost/test" { //nolint:gosec // test credential
		t.Errorf("expected DATABASE_URL, got %q", cfg.DatabaseURL)
	}
	if cfg.SerpAPIKey != "key-123" {
		t.Errorf("expected SERPAPI_KEY=key-123, got %q", cfg.SerpAPIKey)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.ServerPort)
	}
	if cfg.Env != "development" {
		t.Errorf("expected default env=development, got %q", cfg.Env)
	}
	if cfg.FrontendURL != "http://localhost:3000" {
		t.Errorf("expected default frontend URL, got %q", cfg.FrontendURL)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_InvalidServerPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	t.Setenv("SERVER_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SERVER_PORT")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("FRONTEND_URL", "https://app.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerPort != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.ServerPort)
	}
	if cfg.Env != "production" {
		t.Errorf("expected env=production, got %q", cfg.Env)
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("expected frontend URL https://app.example.com, got %q", cfg.FrontendURL)
	}
}
