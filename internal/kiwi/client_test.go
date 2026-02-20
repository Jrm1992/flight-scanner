package kiwi

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
		if r.Header.Get("apikey") != "test-key" {
			t.Errorf("expected apikey header, got %q", r.Header.Get("apikey"))
		}
		if r.URL.Query().Get("fly_from") != "GIG" {
			t.Errorf("expected fly_from=GIG, got %q", r.URL.Query().Get("fly_from"))
		}
		if r.URL.Query().Get("fly_to") != "JFK" {
			t.Errorf("expected fly_to=JFK, got %q", r.URL.Query().Get("fly_to"))
		}

		resp := SearchResponse{
			Data: []Flight{
				{
					ID:             "flight1",
					Price:          450.00,
					Airlines:       []string{"LATAM"},
					FlyFrom:        "GIG",
					FlyTo:          "JFK",
					CityFrom:       "Rio de Janeiro",
					CityTo:         "New York",
					LocalDeparture: "2026-04-01T10:00:00.000Z",
					LocalArrival:   "2026-04-01T18:00:00.000Z",
					DeepLink:       "https://kiwi.com/booking/flight1",
				},
			},
			Currency: "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Override URL to use test server
	rawURL := srv.URL + searchPath + "?fly_from=GIG&fly_to=JFK&date_from=01/04/2026&date_to=30/04/2026&curr=USD"
	results, err := client.doSearch(context.Background(), rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Price != 450.00 {
		t.Errorf("expected price 450, got %f", r.Price)
	}
	if r.Airline != "LATAM" {
		t.Errorf("expected airline LATAM, got %q", r.Airline)
	}
	if r.FlyFrom != "GIG" {
		t.Errorf("expected FlyFrom GIG, got %q", r.FlyFrom)
	}
	if r.FlyTo != "JFK" {
		t.Errorf("expected FlyTo JFK, got %q", r.FlyTo)
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
		resp := SearchResponse{
			Data:     []Flight{{ID: "f1", Price: 300, Airlines: []string{"GOL"}, FlyFrom: "GRU", FlyTo: "MIA"}},
			Currency: "USD",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	// Use Search method with overridden baseURL via direct call
	rawURL := srv.URL + searchPath + "?fly_from=GRU&fly_to=MIA&date_from=01/04/2026&date_to=30/04/2026&curr=USD"

	// Manually test retry logic
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
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + searchPath + "?fly_from=XXX&fly_to=YYY&date_from=01/04/2026&date_to=30/04/2026&curr=USD"
	_, err := client.doSearch(context.Background(), rawURL)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if isRetryable(err) {
		t.Error("400 error should not be retryable")
	}
	if attempts.Load() != 1 {
		t.Errorf("expected 1 attempt (no retry for 4xx), got %d", attempts.Load())
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

	rawURL := srv.URL + searchPath + "?fly_from=GIG&fly_to=JFK&date_from=01/04/2026&date_to=30/04/2026&curr=USD"
	_, err := client.doSearch(context.Background(), rawURL)
	if err == nil {
		t.Fatal("expected error for 429")
	}
	if !isRetryable(err) {
		t.Error("429 should be retryable")
	}
}

func TestBuildSearchURL(t *testing.T) {
	client := NewClient("key")
	params := SearchParams{
		FlyFrom:      "GIG",
		FlyTo:        "JFK",
		DateFrom:     time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		DateTo:       time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		MaxStopovers: 2,
		Limit:        10,
	}

	u, err := client.buildSearchURL(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(u, "fly_from=GIG") {
		t.Errorf("URL missing fly_from: %s", u)
	}
	if !strings.Contains(u, "fly_to=JFK") {
		t.Errorf("URL missing fly_to: %s", u)
	}
	if !strings.Contains(u, "date_from=01%2F04%2F2026") {
		t.Errorf("URL missing or wrong date_from: %s", u)
	}
	if !strings.Contains(u, "curr=USD") {
		t.Errorf("URL missing curr: %s", u)
	}
	if !strings.Contains(u, "max_stopovers=2") {
		t.Errorf("URL missing max_stopovers: %s", u)
	}
	if !strings.Contains(u, "limit=10") {
		t.Errorf("URL missing limit: %s", u)
	}
}

func TestToFlightResults_MultipleAirlines(t *testing.T) {
	flights := []Flight{
		{
			Price:          500,
			Airlines:       []string{"LATAM", "GOL"},
			FlyFrom:        "GRU",
			FlyTo:          "CDG",
			LocalDeparture: "2026-05-01T08:00:00.000Z",
			LocalArrival:   "2026-05-02T06:00:00.000Z",
		},
	}

	results := toFlightResults(flights)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Airline != "LATAM,GOL" {
		t.Errorf("expected airlines joined, got %q", results[0].Airline)
	}
}
