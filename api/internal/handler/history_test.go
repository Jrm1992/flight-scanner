package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jose/flight-scanner/internal/models"
)

type mockPriceHistoryRepo struct {
	history []models.PriceHistory
	stats   *models.PriceStats
	prices  map[string]models.PriceHistory
	err     error
}

func (m *mockPriceHistoryRepo) GetByRoute(_ context.Context, _ string, _ int) ([]models.PriceHistory, error) {
	return m.history, m.err
}

func (m *mockPriceHistoryRepo) GetStats(_ context.Context, _ string, _ int) (*models.PriceStats, error) {
	if m.stats != nil {
		return m.stats, m.err
	}
	return &models.PriceStats{}, m.err
}

func (m *mockPriceHistoryRepo) GetLatestPrices(_ context.Context, _ []string) (map[string]models.PriceHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.prices != nil {
		return m.prices, nil
	}
	return map[string]models.PriceHistory{}, nil
}

func TestHistoryHandler_GetHistory(t *testing.T) {
	tests := []struct {
		name     string
		routeID  string
		query    string
		repo     *mockPriceHistoryRepo
		wantCode int
		wantDays int
	}{
		{
			name:    "default days",
			routeID: "r-1",
			query:   "",
			repo: &mockPriceHistoryRepo{
				history: []models.PriceHistory{
					{ID: "ph-1", RouteID: "r-1", MinPrice: 200, MaxPrice: 400, AvgPrice: 300, CheckedAt: time.Now()},
				},
				stats: &models.PriceStats{MinPrice: 200, MaxPrice: 400, AvgPrice: 300},
			},
			wantCode: http.StatusOK,
			wantDays: 30,
		},
		{
			name:    "custom days",
			routeID: "r-1",
			query:   "?days=7",
			repo: &mockPriceHistoryRepo{
				history: []models.PriceHistory{},
				stats:   &models.PriceStats{},
			},
			wantCode: http.StatusOK,
			wantDays: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistoryHandler(tt.repo)
			req := httptest.NewRequest(http.MethodGet, "/api/routes/"+tt.routeID+"/history"+tt.query, nil)
			req.SetPathValue("id", tt.routeID)
			w := httptest.NewRecorder()
			h.GetHistory(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			var body struct {
				Count int `json:"count"`
				Days  int `json:"days"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("json unmarshal: %v", err)
			}
			if body.Days != tt.wantDays {
				t.Fatalf("expected days=%d, got %d", tt.wantDays, body.Days)
			}
		})
	}
}

func TestHistoryHandler_Export(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantCode    int
		wantCT      string
		wantContain string
	}{
		{
			name:        "CSV export",
			query:       "",
			wantCode:    http.StatusOK,
			wantCT:      "text/csv",
			wantContain: "LATAM",
		},
		{
			name:        "JSON export",
			query:       "?format=json",
			wantCode:    http.StatusOK,
			wantCT:      "application/json",
			wantContain: ".json",
		},
	}

	repo := &mockPriceHistoryRepo{
		history: []models.PriceHistory{
			{ID: "ph-1", RouteID: "r-1234567890", MinPrice: 200, MaxPrice: 400, AvgPrice: 300, Airline: "LATAM", CheckedAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistoryHandler(repo)
			req := httptest.NewRequest(http.MethodGet, "/api/routes/r-1234567890/history/export"+tt.query, nil)
			req.SetPathValue("id", "r-1234567890")
			w := httptest.NewRecorder()
			h.Export(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			ct := w.Header().Get("Content-Type")
			if ct != tt.wantCT {
				t.Fatalf("expected content-type %s, got %s", tt.wantCT, ct)
			}

			if tt.wantCT == "text/csv" {
				if !strings.Contains(w.Body.String(), tt.wantContain) {
					t.Fatalf("expected body to contain %q", tt.wantContain)
				}
			} else {
				disp := w.Header().Get("Content-Disposition")
				if !strings.Contains(disp, tt.wantContain) {
					t.Fatalf("expected Content-Disposition to contain %q, got %s", tt.wantContain, disp)
				}
			}
		})
	}
}
