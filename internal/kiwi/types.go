package kiwi

import "time"

// SearchParams holds the query parameters for a Kiwi flight search.
type SearchParams struct {
	FlyFrom  string // IATA code (e.g. "GIG")
	FlyTo    string // IATA code (e.g. "JFK")
	DateFrom time.Time
	DateTo   time.Time
	Currency string // e.g. "USD", "BRL"
	MaxStopovers int
	Limit    int // max results to return (0 = API default)
}

// SearchResponse is the top-level Kiwi /v2/search response.
type SearchResponse struct {
	Data     []Flight `json:"data"`
	Currency string   `json:"currency"`
}

// Flight represents a single flight option returned by the Kiwi API.
type Flight struct {
	ID           string   `json:"id"`
	Price        float64  `json:"price"`
	Airlines     []string `json:"airlines"`
	CityFrom     string   `json:"cityFrom"`
	CityTo       string   `json:"cityTo"`
	FlyFrom      string   `json:"flyFrom"`
	FlyTo        string   `json:"flyTo"`
	LocalDeparture string `json:"local_departure"`
	LocalArrival   string `json:"local_arrival"`
	Route        []Leg    `json:"route"`
	DeepLink     string   `json:"deep_link"`
}

// Leg represents a single segment (leg) within a flight itinerary.
type Leg struct {
	ID             string  `json:"id"`
	FlyFrom        string  `json:"flyFrom"`
	FlyTo          string  `json:"flyTo"`
	CityFrom       string  `json:"cityFrom"`
	CityTo         string  `json:"cityTo"`
	Airline        string  `json:"airline"`
	FlightNo       int     `json:"flight_no"`
	LocalDeparture string  `json:"local_departure"`
	LocalArrival   string  `json:"local_arrival"`
}

// FlightResult is our internal simplified representation of a search result,
// used by the rest of the application.
type FlightResult struct {
	Price    float64
	Airline  string
	FlyFrom  string
	FlyTo    string
	Departure time.Time
	Arrival   time.Time
	DeepLink string
}
