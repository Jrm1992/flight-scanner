package flightapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSearch_FullFlow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SerpResponse{
			BestFlights: []FlightGroup{
				{
					Flights: []FlightLeg{
						{
							DepartureAirport: Airport{Name: "Rio Galeao", ID: "GIG", Time: "2026-04-01 10:00"},
							ArrivalAirport:   Airport{Name: "Santiago", ID: "SCL", Time: "2026-04-01 16:30"},
							Duration:         390,
							Airline:          "LATAM",
							FlightNumber:     "LA8090",
						},
					},
					TotalDuration: 390,
					Price:         320,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Override baseURL by using doSearch directly with test server URL
	params := SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
	}

	u := srv.URL + "?" + "engine=google_flights&departure_id=" + params.DepartureID + "&arrival_id=" + params.ArrivalID

	result, err := client.doSearch(context.Background(), u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Flights) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Flights))
	}
	if result.Flights[0].Price != 320 {
		t.Errorf("expected price 320, got %f", result.Flights[0].Price)
	}
	if result.Flights[0].Airline != "LATAM" {
		t.Errorf("expected LATAM, got %q", result.Flights[0].Airline)
	}
}

func TestSearch_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.doSearch(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
