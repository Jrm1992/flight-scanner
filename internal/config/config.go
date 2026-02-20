package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	KiwiAPIKey  string
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

	return &Config{
		DatabaseURL: dbURL,
		KiwiAPIKey:  os.Getenv("KIWI_API_KEY"),
		ServerPort:  port,
		Env:         env,
		FrontendURL: frontendURL,
	}, nil
}
