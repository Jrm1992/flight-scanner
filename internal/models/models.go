package models

import "time"

// Route represents a flight route being monitored for price changes.
type Route struct {
	ID                    string    `json:"id"`
	Origin                string    `json:"origin"`
	Destination           string    `json:"destination"`
	AlertPrice            float64   `json:"alert_price"`
	CheckFrequencyMinutes int       `json:"check_frequency_minutes"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
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
	AlertPrice            float64 `json:"alert_price"`
	CheckFrequencyMinutes int     `json:"check_frequency_minutes"`
}

// UpdateRouteRequest is the payload for updating route configuration.
type UpdateRouteRequest struct {
	AlertPrice            *float64 `json:"alert_price,omitempty"`
	CheckFrequencyMinutes *int     `json:"check_frequency_minutes,omitempty"`
}
