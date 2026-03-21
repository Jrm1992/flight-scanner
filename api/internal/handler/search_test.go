package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jose/flight-scanner/internal/flightapi"
)

type mockFlightSearcher struct {
	results              []flightapi.FlightResult
	err                  error
	autocompleteResults  []flightapi.AutocompleteResult
	autocompleteErr      error
}

func (m *mockFlightSearcher) Search(_ context.Context, _ flightapi.SearchParams) ([]flightapi.FlightResult, error) {
	return m.results, m.err
}

func (m *mockFlightSearcher) Autocomplete(_ context.Context, _ string) ([]flightapi.AutocompleteResult, error) {
	return m.autocompleteResults, m.autocompleteErr
}

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		mock     *mockFlightSearcher
		wantCode int
	}{
		{
			name:    "success",
			payload: `{"origin":"GIG","destination":"SCL","date":"2026-05-01"}`,
			mock: &mockFlightSearcher{
				results: []flightapi.FlightResult{
					{Price: 299, Airline: "LATAM", DepartureCode: "GIG", ArrivalCode: "SCL"},
				},
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid JSON",
			payload:  "{bad",
			mock:     &mockFlightSearcher{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid IATA",
			payload:  `{"origin":"X","destination":"SCL"}`,
			mock:     &mockFlightSearcher{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "search error",
			payload:  `{"origin":"GIG","destination":"SCL"}`,
			mock:     &mockFlightSearcher{err: errors.New("api down")},
			wantCode: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHandler(tt.mock)
			req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(tt.payload))
			w := httptest.NewRecorder()
			h.Search(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			if tt.wantCode == http.StatusOK {
				var resp map[string]json.RawMessage
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("json unmarshal: %v", err)
				}
				if string(resp["count"]) != "1" {
					t.Fatalf("expected count 1, got %s", resp["count"])
				}
			}
		})
	}
}

func TestSearchHandler_InvalidDate(t *testing.T) {
	h := NewSearchHandler(&mockFlightSearcher{})
	payload := `{"origin":"GIG","destination":"SCL","date":"not-a-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearchHandler_InvalidReturnDate(t *testing.T) {
	h := NewSearchHandler(&mockFlightSearcher{})
	payload := `{"origin":"GIG","destination":"SCL","date":"2026-05-01","return_date":"bad-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearchHandler_WithReturnDate(t *testing.T) {
	mock := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 500, Airline: "LATAM", DepartureCode: "GIG", ArrivalCode: "SCL"},
		},
	}
	h := NewSearchHandler(mock)
	payload := `{"origin":"GIG","destination":"SCL","date":"2026-05-01","return_date":"2026-05-15"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearchHandler_DefaultCurrency(t *testing.T) {
	mock := &mockFlightSearcher{
		results: []flightapi.FlightResult{},
	}
	h := NewSearchHandler(mock)
	// No currency specified — should default to USD
	payload := `{"origin":"GIG","destination":"SCL","date":"2026-05-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if resp.Currency != "USD" {
		t.Fatalf("expected currency USD, got %q", resp.Currency)
	}
}

func TestAutocompleteHandler(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		mock     *mockFlightSearcher
		wantCode int
	}{
		{
			name:  "success",
			query: "san",
			mock: &mockFlightSearcher{
				autocompleteResults: []flightapi.AutocompleteResult{
					{Code: "SAN", Name: "San Diego Intl", City: "San Diego"},
					{Code: "SCL", Name: "Santiago Intl", City: "Santiago"},
				},
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "empty query returns empty array",
			query:    "",
			mock:     &mockFlightSearcher{},
			wantCode: http.StatusOK,
		},
		{
			name:     "api error",
			query:    "test",
			mock:     &mockFlightSearcher{autocompleteErr: errors.New("api down")},
			wantCode: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHandler(tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/api/search/airports?q="+tt.query, nil)
			w := httptest.NewRecorder()
			h.Autocomplete(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			if tt.wantCode == http.StatusOK && tt.query != "" {
				var resp []flightapi.AutocompleteResult
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("json unmarshal: %v", err)
				}
				if len(resp) != 2 {
					t.Fatalf("expected 2 results, got %d", len(resp))
				}
			}
		})
	}
}

func TestSearchHandler_NoDate(t *testing.T) {
	mock := &mockFlightSearcher{
		results: []flightapi.FlightResult{},
	}
	h := NewSearchHandler(mock)
	// No date — should default to tomorrow
	payload := `{"origin":"GIG","destination":"SCL"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search/flights", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
