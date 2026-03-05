package models

import "time"

// User represents a registered user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Route represents a flight route being monitored for price changes.
type Route struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Origin                string    `json:"origin"`
	Destination           string    `json:"destination"`
	DepartureDate         string    `json:"departure_date"`
	ReturnDate            *string   `json:"return_date,omitempty"`
	AlertPrice            float64   `json:"alert_price"`
	CheckFrequencyMinutes int       `json:"check_frequency_minutes"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// RouteWithPrice extends Route with the latest price information.
type RouteWithPrice struct {
	Route
	CurrentPrice *float64   `json:"current_price,omitempty"`
	LastCheckAt  *time.Time `json:"last_check_at,omitempty"`
	PriceTrend   string     `json:"price_trend,omitempty"`
}

// PriceHistory stores a price snapshot for a monitored route.
type PriceHistory struct {
	ID        string    `json:"id"`
	RouteID   string    `json:"route_id"`
	MinPrice  float64   `json:"min_price"`
	MaxPrice  float64   `json:"max_price"`
	AvgPrice  float64   `json:"avg_price"`
	Airline   string    `json:"airline,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// Alert is created when a flight price drops below the configured alert threshold.
type Alert struct {
	ID             string     `json:"id"`
	RouteID        string     `json:"route_id"`
	AlertPrice     float64    `json:"alert_price"`
	TriggeredPrice float64    `json:"triggered_price"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	Notified       bool       `json:"notified"`
	NotifiedAt     *time.Time `json:"notified_at,omitempty"`
}

// CreateRouteRequest is the payload for creating a new monitored route.
type CreateRouteRequest struct {
	Origin                string  `json:"origin"`
	Destination           string  `json:"destination"`
	DepartureDate         string  `json:"departure_date"`
	ReturnDate            *string `json:"return_date,omitempty"`
	AlertPrice            float64 `json:"alert_price"`
	CheckFrequencyMinutes int     `json:"check_frequency_minutes"`
}

// UpdateRouteRequest is the payload for updating route configuration.
type UpdateRouteRequest struct {
	AlertPrice            *float64 `json:"alert_price,omitempty"`
	CheckFrequencyMinutes *int     `json:"check_frequency_minutes,omitempty"`
}

// PriceStats holds aggregated price statistics for a route over a period.
type PriceStats struct {
	MinPrice float64   `json:"min_price"`
	MaxPrice float64   `json:"max_price"`
	AvgPrice float64   `json:"avg_price"`
	Since    time.Time `json:"since"`
}
