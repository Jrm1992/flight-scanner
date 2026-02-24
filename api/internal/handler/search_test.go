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
	results []flightapi.FlightResult
	err     error
}

func (m *mockFlightSearcher) Search(_ context.Context, _ flightapi.SearchParams) ([]flightapi.FlightResult, error) {
	return m.results, m.err
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
