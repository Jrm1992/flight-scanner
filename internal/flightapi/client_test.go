package flightapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSearch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("engine") != "google_flights" {
			t.Errorf("expected engine=google_flights, got %q", r.URL.Query().Get("engine"))
		}
		if r.URL.Query().Get("departure_id") != "GIG" {
			t.Errorf("expected departure_id=GIG, got %q", r.URL.Query().Get("departure_id"))
		}
		if r.URL.Query().Get("arrival_id") != "SCL" {
			t.Errorf("expected arrival_id=SCL, got %q", r.URL.Query().Get("arrival_id"))
		}
		if r.URL.Query().Get("type") != "2" {
			t.Errorf("expected type=2 (one way), got %q", r.URL.Query().Get("type"))
		}

		resp := SerpResponse{
			BestFlights: []FlightGroup{
				{
					Flights: []FlightLeg{
						{
							DepartureAirport: Airport{Name: "Rio Galeão", ID: "GIG", Time: "2026-04-01 10:00"},
							ArrivalAirport:   Airport{Name: "Santiago", ID: "SCL", Time: "2026-04-01 16:30"},
							Duration:         390,
							Airline:          "LATAM",
							FlightNumber:     "LA8090",
						},
					},
					TotalDuration: 390,
					Price:         320,
					Type:          "One way",
				},
			},
			OtherFlights: []FlightGroup{
				{
					Flights: []FlightLeg{
						{
							DepartureAirport: Airport{Name: "Rio Galeão", ID: "GIG", Time: "2026-04-01 22:00"},
							ArrivalAirport:   Airport{Name: "Ezeiza", ID: "EZE", Time: "2026-04-02 02:00"},
							Duration:         240,
							Airline:          "Aerolíneas",
							FlightNumber:     "AR1234",
						},
						{
							DepartureAirport: Airport{Name: "Ezeiza", ID: "EZE", Time: "2026-04-02 06:00"},
							ArrivalAirport:   Airport{Name: "Santiago", ID: "SCL", Time: "2026-04-02 08:00"},
							Duration:         120,
							Airline:          "Aerolíneas",
							FlightNumber:     "AR5678",
						},
					},
					Layovers: []Layover{
						{Duration: 240, Name: "Ezeiza", ID: "EZE"},
					},
					TotalDuration: 600,
					Price:         250,
					Type:          "One way",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights&departure_id=GIG&arrival_id=SCL&outbound_date=2026-04-01&type=2&currency=USD&adults=1&travel_class=1&hl=en&api_key=test-key"
	results, err := client.doSearch(context.Background(), rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Best flight: direct LATAM
	r0 := results[0]
	if r0.Price != 320 {
		t.Errorf("expected price 320, got %f", r0.Price)
	}
	if r0.Airline != "LATAM" {
		t.Errorf("expected airline LATAM, got %q", r0.Airline)
	}
	if r0.Stops != 0 {
		t.Errorf("expected 0 stops, got %d", r0.Stops)
	}
	if r0.DepartureCode != "GIG" {
		t.Errorf("expected departure GIG, got %q", r0.DepartureCode)
	}
	if r0.Duration != 390 {
		t.Errorf("expected duration 390, got %d", r0.Duration)
	}

	// Other flight: 1 stop via EZE
	r1 := results[1]
	if r1.Price != 250 {
		t.Errorf("expected price 250, got %f", r1.Price)
	}
	if r1.Stops != 1 {
		t.Errorf("expected 1 stop, got %d", r1.Stops)
	}
	if r1.ArrivalCode != "SCL" {
		t.Errorf("expected arrival SCL, got %q", r1.ArrivalCode)
	}
}

func TestSearch_RetryOnServerError(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}
		resp := SerpResponse{
			BestFlights: []FlightGroup{
				{
					Flights:       []FlightLeg{{DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"}, ArrivalAirport: Airport{ID: "SCL", Time: "2026-04-01 16:00"}, Airline: "GOL"}},
					TotalDuration: 360,
					Price:         300,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights&departure_id=GIG&arrival_id=SCL"

	var lastErr error
	var results []FlightResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err := client.doSearch(context.Background(), rawURL)
		if err == nil {
			results = res
			break
		}
		lastErr = err
		if !isRetryable(err) {
			break
		}
	}

	if results == nil {
		t.Fatalf("expected success after retries, got: %v", lastErr)
	}
	if len(results) != 1 || results[0].Price != 300 {
		t.Errorf("unexpected result: %+v", results)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestSearch_NoRetryOn4xx(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid API key"}`))
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "bad-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights"
	_, err := client.doSearch(context.Background(), rawURL)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if isRetryable(err) {
		t.Error("400 error should not be retryable")
	}
	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts.Load())
	}
}

func TestSearch_RateLimitRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights"
	_, err := client.doSearch(context.Background(), rawURL)
	if err == nil {
		t.Fatal("expected error for 429")
	}
	if !isRetryable(err) {
		t.Error("429 should be retryable")
	}
}

func TestBuildSearchURL(t *testing.T) {
	client := NewClient("my-key")
	outbound := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	ret := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	params := SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: outbound,
		ReturnDate:   &ret,
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
		Stops:        1,
		MaxPrice:     500,
	}

	u := client.buildSearchURL(params)

	checks := map[string]string{
		"engine":        "google_flights",
		"departure_id":  "GIG",
		"arrival_id":    "SCL",
		"outbound_date": "2026-04-01",
		"return_date":   "2026-04-15",
		"type":          "1", // round trip
		"currency":      "USD",
		"stops":         "1",
		"max_price":     "500",
		"api_key":       "my-key",
	}

	for key, expected := range checks {
		if !strings.Contains(u, key+"="+expected) {
			t.Errorf("URL missing %s=%s: %s", key, expected, u)
		}
	}
}

func TestBuildSearchURL_OneWay(t *testing.T) {
	client := NewClient("key")
	params := SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "BRL",
	}

	u := client.buildSearchURL(params)

	if !strings.Contains(u, "type=2") {
		t.Errorf("expected type=2 for one way, got: %s", u)
	}
	if strings.Contains(u, "return_date") {
		t.Errorf("one way should not have return_date: %s", u)
	}
}

func TestToFlightResults_MultiLeg(t *testing.T) {
	groups := []FlightGroup{
		{
			Flights: []FlightLeg{
				{
					DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"},
					ArrivalAirport:   Airport{ID: "EZE", Time: "2026-04-01 14:00"},
					Airline:          "LATAM",
					FlightNumber:     "LA800",
				},
				{
					DepartureAirport: Airport{ID: "EZE", Time: "2026-04-01 17:00"},
					ArrivalAirport:   Airport{ID: "SCL", Time: "2026-04-01 19:00"},
					Airline:          "LATAM",
					FlightNumber:     "LA900",
				},
			},
			Layovers:      []Layover{{Duration: 180, Name: "Ezeiza", ID: "EZE"}},
			TotalDuration: 540,
			Price:         280,
		},
	}

	results := toFlightResults(groups)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.DepartureCode != "GIG" {
		t.Errorf("expected dep GIG, got %q", r.DepartureCode)
	}
	if r.ArrivalCode != "SCL" {
		t.Errorf("expected arr SCL, got %q", r.ArrivalCode)
	}
	if r.Stops != 1 {
		t.Errorf("expected 1 stop, got %d", r.Stops)
	}
	if r.Airline != "LATAM" {
		t.Errorf("expected LATAM (same airline both legs), got %q", r.Airline)
	}
}
