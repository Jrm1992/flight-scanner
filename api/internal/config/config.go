package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	SerpAPIKey  string
	JWTSecret   string
	ServerPort  int
	Env         string
	FrontendURL string
}

// Load reads configuration from environment variables and returns a Config.
// It returns an error if required variables are missing.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := 8080
	if p := os.Getenv("SERVER_PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
		}
		if parsed < 1 || parsed > 65535 {
			return nil, fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", parsed)
		}
		port = parsed
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	if os.Getenv("SERPAPI_KEY") == "" {
		slog.Warn("SERPAPI_KEY is not set; flight searches will fail")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return &Config{
		DatabaseURL: dbURL,
		SerpAPIKey:  os.Getenv("SERPAPI_KEY"),
		JWTSecret:   jwtSecret,
		ServerPort:  port,
		Env:         env,
		FrontendURL: frontendURL,
	}, nil
}
