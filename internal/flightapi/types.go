package flightapi

import "time"

// SearchParams holds query parameters for a Google Flights search via SerpApi.
type SearchParams struct {
	DepartureID string // IATA code (e.g. "GIG")
	ArrivalID   string // IATA code (e.g. "SCL")
	OutboundDate time.Time
	ReturnDate   *time.Time // nil = one way
	Currency     string     // e.g. "USD", "BRL"
	Adults       int
	TravelClass  int // 1=economy, 2=premium, 3=business, 4=first
	Stops        int // 0=any, 1=nonstop, 2=1stop, 3=2stops
	MaxPrice     int // 0 = no limit
}

// SerpResponse is the top-level SerpApi Google Flights response.
type SerpResponse struct {
	BestFlights   []FlightGroup  `json:"best_flights"`
	OtherFlights  []FlightGroup  `json:"other_flights"`
	PriceInsights *PriceInsights `json:"price_insights"`
}

// FlightGroup represents a flight option (may have multiple legs/layovers).
type FlightGroup struct {
	Flights       []FlightLeg `json:"flights"`
	Layovers      []Layover   `json:"layovers"`
	TotalDuration int         `json:"total_duration"` // minutes
	Price         int         `json:"price"`
	Type          string      `json:"type"`
}

// FlightLeg is a single segment within a flight group.
type FlightLeg struct {
	DepartureAirport Airport `json:"departure_airport"`
	ArrivalAirport   Airport `json:"arrival_airport"`
	Duration         int     `json:"duration"` // minutes
	Airplane         string  `json:"airplane"`
	Airline          string  `json:"airline"`
	AirlineLogo      string  `json:"airline_logo"`
	FlightNumber     string  `json:"flight_number"`
	TravelClass      string  `json:"travel_class"`
	Legroom          string  `json:"legroom"`
}

// Airport contains airport info from the response.
type Airport struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Time string `json:"time"` // "2026-04-01 10:00"
}

// Layover represents a connection between legs.
type Layover struct {
	Duration  int    `json:"duration"` // minutes
	Name      string `json:"name"`
	ID        string `json:"id"`
	Overnight bool   `json:"overnight"`
}

// PriceInsights contains pricing analysis from Google Flights.
type PriceInsights struct {
	LowestPrice       int        `json:"lowest_price"`
	PriceLevel        string     `json:"price_level"`
	TypicalPriceRange [2]int     `json:"typical_price_range"`
	PriceHistory      [][2]int64 `json:"price_history"` // [timestamp, price]
}

// FlightResult is our internal simplified representation used by the rest of the app.
type FlightResult struct {
	Price         float64
	Airline       string
	FlightNumber  string
	DepartureCode string
	ArrivalCode   string
	Departure     time.Time
	Arrival       time.Time
	Duration      int // total minutes
	Stops         int
}
