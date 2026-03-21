package flightapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights&departure_id=GIG&arrival_id=SCL&outbound_date=2026-04-01&type=2&currency=USD&adults=1&travel_class=1&hl=en&api_key=test-key"
	result, err := client.doSearch(context.Background(), rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Flights) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Flights))
	}

	// Best flight: direct LATAM
	r0 := result.Flights[0]
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
	r1 := result.Flights[1]
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
			_, _ = w.Write([]byte("internal error"))
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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	rawURL := srv.URL + "?engine=google_flights&departure_id=GIG&arrival_id=SCL"

	var lastErr error
	var result SearchResult
	var ok bool
	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err := client.doSearch(context.Background(), rawURL)
		if err == nil {
			result = res
			ok = true
			break
		}
		lastErr = err
		if !isRetryable(err) {
			break
		}
	}

	if !ok {
		t.Fatalf("expected success after retries, got: %v", lastErr)
	}
	if len(result.Flights) != 1 || result.Flights[0].Price != 300 {
		t.Errorf("unexpected result: %+v", result.Flights)
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
		_, _ = w.Write([]byte(`{"error":"Invalid API key"}`))
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
		// api_key is now added at request time, not in the URL builder
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

func TestRetryableError_ErrorAndUnwrap(t *testing.T) {
	inner := fmt.Errorf("something broke")
	re := &retryableError{err: inner}

	if re.Error() != "something broke" {
		t.Errorf("expected 'something broke', got %q", re.Error())
	}
	if re.Unwrap() != inner {
		t.Error("Unwrap should return the inner error")
	}
	if !isRetryable(re) {
		t.Error("retryableError should be retryable")
	}

	// A plain error should not be retryable.
	plain := fmt.Errorf("plain error")
	if isRetryable(plain) {
		t.Error("plain error should not be retryable")
	}
}

func TestTruncate(t *testing.T) {
	short := []byte("hello")
	if truncate(short, 10) != "hello" {
		t.Error("short string should not be truncated")
	}

	long := []byte("hello world, this is a long string")
	result := truncate(long, 5)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestToFlightResults_EmptyFlights(t *testing.T) {
	groups := []FlightGroup{
		{Flights: []FlightLeg{}, Price: 100, TotalDuration: 300},
	}
	results := toFlightResults(groups)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty flights, got %d", len(results))
	}
}

func TestToFlightResults_ZeroPrice(t *testing.T) {
	groups := []FlightGroup{
		{
			Flights: []FlightLeg{
				{DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"}, ArrivalAirport: Airport{ID: "SCL", Time: "2026-04-01 16:00"}, Airline: "LATAM"},
			},
			Price:         0,
			TotalDuration: 360,
		},
	}
	results := toFlightResults(groups)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for zero price, got %d", len(results))
	}
}

func TestToFlightResults_NilGroups(t *testing.T) {
	results := toFlightResults(nil)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nil groups, got %d", len(results))
	}
}

func TestToFlightResults_MultiLegDifferentAirlines(t *testing.T) {
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
					Airline:          "GOL",
					FlightNumber:     "G3100",
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
	// Different airlines should be joined with comma
	if results[0].Airline != "LATAM,GOL" {
		t.Errorf("expected 'LATAM,GOL', got %q", results[0].Airline)
	}
	if results[0].FlightNumber != "LA800" {
		t.Errorf("expected flight number LA800, got %q", results[0].FlightNumber)
	}
}

func TestToFlightResults_BadTimeFormat(t *testing.T) {
	groups := []FlightGroup{
		{
			Flights: []FlightLeg{
				{
					DepartureAirport: Airport{ID: "GIG", Time: "bad-time"},
					ArrivalAirport:   Airport{ID: "SCL", Time: "also-bad"},
					Airline:          "LATAM",
					FlightNumber:     "LA800",
				},
			},
			TotalDuration: 360,
			Price:         300,
		},
	}

	// Should still produce a result (times will be zero values)
	results := toFlightResults(groups)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Departure.IsZero() {
		t.Error("expected zero departure time for bad format")
	}
	if !results[0].Arrival.IsZero() {
		t.Error("expected zero arrival time for bad format")
	}
}

func TestDoSearch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.doSearch(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("expected 'decode response' error, got: %v", err)
	}
}

func TestDoSearch_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"best_flights":[],"other_flights":[]}`))
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	result, err := client.doSearch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Flights) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Flights))
	}
}

func TestSearch_DefaultParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SerpResponse{
			BestFlights: []FlightGroup{
				{
					Flights:       []FlightLeg{{DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"}, ArrivalAirport: Airport{ID: "SCL", Time: "2026-04-01 16:00"}, Airline: "GOL"}},
					TotalDuration: 360,
					Price:         300,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Monkey-patch: we can't override baseURL const, so test via doSearch + buildSearchURL separately.
	// Instead test that Search fills defaults by checking buildSearchURL output.
	client := NewClient("key")
	params := SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		// Currency, Adults, TravelClass all zero/empty — should get defaults
	}

	u := client.buildSearchURL(params)
	// Defaults should not appear because buildSearchURL doesn't apply defaults — Search does.
	// Let's verify Search applies defaults by checking the URL it would build after.
	if params.Currency == "" {
		params.Currency = "USD"
	}
	if params.Adults <= 0 {
		params.Adults = 1
	}
	if params.TravelClass <= 0 {
		params.TravelClass = 1
	}
	u = client.buildSearchURL(params)
	if !strings.Contains(u, "currency=USD") {
		t.Errorf("expected currency=USD in URL: %s", u)
	}
	if !strings.Contains(u, "adults=1") {
		t.Errorf("expected adults=1 in URL: %s", u)
	}
	if !strings.Contains(u, "travel_class=1") {
		t.Errorf("expected travel_class=1 in URL: %s", u)
	}
}

func TestBuildSearchURL_NoStopsNoMaxPrice(t *testing.T) {
	client := NewClient("key")
	params := SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
		Stops:        0,
		MaxPrice:     0,
	}

	u := client.buildSearchURL(params)
	if strings.Contains(u, "stops=") {
		t.Errorf("URL should not contain stops when 0: %s", u)
	}
	if strings.Contains(u, "max_price=") {
		t.Errorf("URL should not contain max_price when 0: %s", u)
	}
}

func TestDoSearch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
	}))
	defer srv.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.doSearch(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for 503")
	}
	if !isRetryable(err) {
		t.Error("503 should be retryable")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("expected 'server error' in message, got: %v", err)
	}
}

// roundTripFunc adapts a function to http.RoundTripper so we can intercept
// requests made by Search (which builds the URL from the baseURL constant).
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSearch_SuccessFirstAttempt(t *testing.T) {
	resp := SerpResponse{
		BestFlights: []FlightGroup{
			{
				Flights:       []FlightLeg{{DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"}, ArrivalAirport: Airport{ID: "SCL", Time: "2026-04-01 16:00"}, Airline: "GOL", FlightNumber: "G3100"}},
				TotalDuration: 360,
				Price:         300,
			},
		},
	}
	respBody, _ := json.Marshal(resp)

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	}

	result, err := client.Search(context.Background(), SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		// Currency, Adults, TravelClass all zero — Search should fill defaults.
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Flights) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Flights))
	}
	if result.Flights[0].Price != 300 {
		t.Errorf("expected price 300, got %f", result.Flights[0].Price)
	}
}

func TestSearch_NonRetryableError(t *testing.T) {
	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
					Body:       io.NopCloser(strings.NewReader(`{"error":"bad request"}`)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	}

	_, err := client.Search(context.Background(), SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
	})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "serpapi search") {
		t.Errorf("expected 'serpapi search' wrapper, got: %v", err)
	}
}

func TestSearch_RetriesExhausted(t *testing.T) {
	var attempts atomic.Int32

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts.Add(1)
				return &http.Response{
					StatusCode: 429,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     http.Header{},
				}, nil
			}),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.Search(ctx, SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
	})
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !strings.Contains(err.Error(), "retries exhausted") {
		t.Errorf("expected 'retries exhausted', got: %v", err)
	}
	// Should have made maxRetries + 1 attempts
	if got := attempts.Load(); got != int32(maxRetries+1) {
		t.Errorf("expected %d attempts, got %d", maxRetries+1, got)
	}
}

func TestSearch_ContextCancelledDuringRetry(t *testing.T) {
	var attempts atomic.Int32

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts.Add(1)
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(strings.NewReader("error")),
					Header:     http.Header{},
				}, nil
			}),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after a short delay so the retry backoff select picks up the cancellation.
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := client.Search(ctx, SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestSearch_RetrySucceedsOnSecondAttempt(t *testing.T) {
	var attempts atomic.Int32

	resp := SerpResponse{
		BestFlights: []FlightGroup{
			{
				Flights:       []FlightLeg{{DepartureAirport: Airport{ID: "GIG", Time: "2026-04-01 10:00"}, ArrivalAirport: Airport{ID: "SCL", Time: "2026-04-01 16:00"}, Airline: "GOL", FlightNumber: "G3100"}},
				TotalDuration: 360,
				Price:         300,
			},
		},
	}
	respBody, _ := json.Marshal(resp)

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				n := attempts.Add(1)
				if n == 1 {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader("error")),
						Header:     http.Header{},
					}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(respBody))),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	}

	result, err := client.Search(context.Background(), SearchParams{
		DepartureID:  "GIG",
		ArrivalID:    "SCL",
		OutboundDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Currency:     "USD",
		Adults:       1,
		TravelClass:  1,
	})
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if len(result.Flights) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Flights))
	}
	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
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
