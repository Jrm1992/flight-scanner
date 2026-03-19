package handler

import (
	"context"
	"encoding/json"
	"errors"
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

func (m *mockPriceHistoryRepo) GetByRoute(_ context.Context, _, _ string, _ int) ([]models.PriceHistory, error) {
	return m.history, m.err
}

func (m *mockPriceHistoryRepo) GetStats(_ context.Context, _, _ string, _ int) (*models.PriceStats, error) {
	if m.stats != nil {
		return m.stats, m.err
	}
	return &models.PriceStats{}, m.err
}

func (m *mockPriceHistoryRepo) GetLatestPrices(_ context.Context, _ string, _ []string) (map[string]models.PriceHistory, error) {
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
			req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/"+tt.routeID+"/history"+tt.query, nil))
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
			req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-1234567890/history/export"+tt.query, nil))
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

func TestHistoryHandler_GetHistory_MissingID(t *testing.T) {
	h := NewHistoryHandler(&mockPriceHistoryRepo{})
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes//history", nil))
	w := httptest.NewRecorder()
	h.GetHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_GetHistory_RepoError(t *testing.T) {
	repo := &mockPriceHistoryRepo{err: errors.New("db error")}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-1/history", nil))
	req.SetPathValue("id", "r-1")
	w := httptest.NewRecorder()
	h.GetHistory(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// mockPriceHistoryRepoStatsErr returns success for GetByRoute but error for GetStats.
type mockPriceHistoryRepoStatsErr struct {
	mockPriceHistoryRepo
	statsErr error
}

func (m *mockPriceHistoryRepoStatsErr) GetStats(_ context.Context, _, _ string, _ int) (*models.PriceStats, error) {
	return nil, m.statsErr
}

func TestHistoryHandler_GetHistory_StatsError(t *testing.T) {
	repo := &mockPriceHistoryRepoStatsErr{
		mockPriceHistoryRepo: mockPriceHistoryRepo{
			history: []models.PriceHistory{},
		},
		statsErr: errors.New("stats error"),
	}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-1/history", nil))
	req.SetPathValue("id", "r-1")
	w := httptest.NewRecorder()
	h.GetHistory(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_GetHistory_InvalidDays(t *testing.T) {
	repo := &mockPriceHistoryRepo{
		history: []models.PriceHistory{},
		stats:   &models.PriceStats{},
	}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-1/history?days=abc", nil))
	req.SetPathValue("id", "r-1")
	w := httptest.NewRecorder()
	h.GetHistory(w, req)

	// Should fall back to default 30 days
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Days int `json:"days"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if body.Days != 30 {
		t.Fatalf("expected days=30 for invalid input, got %d", body.Days)
	}
}

func TestHistoryHandler_Export_MissingID(t *testing.T) {
	h := NewHistoryHandler(&mockPriceHistoryRepo{})
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes//history/export", nil))
	w := httptest.NewRecorder()
	h.Export(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_Export_RepoError(t *testing.T) {
	repo := &mockPriceHistoryRepo{err: errors.New("db error")}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-1/history/export", nil))
	req.SetPathValue("id", "r-1")
	w := httptest.NewRecorder()
	h.Export(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_Export_ShortRouteID(t *testing.T) {
	repo := &mockPriceHistoryRepo{
		history: []models.PriceHistory{
			{ID: "ph-1", RouteID: "short", MinPrice: 200, MaxPrice: 400, AvgPrice: 300, Airline: "LATAM", CheckedAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)},
		},
	}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/short/history/export", nil))
	req.SetPathValue("id", "short")
	w := httptest.NewRecorder()
	h.Export(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	disp := w.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "short") {
		t.Fatalf("expected short route ID in filename, got %s", disp)
	}
}

func TestHistoryHandler_Export_CustomDays(t *testing.T) {
	repo := &mockPriceHistoryRepo{
		history: []models.PriceHistory{},
	}
	h := NewHistoryHandler(repo)
	req := withUserCtx(httptest.NewRequest(http.MethodGet, "/api/routes/r-12345678/history/export?days=7", nil))
	req.SetPathValue("id", "r-12345678")
	w := httptest.NewRecorder()
	h.Export(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	disp := w.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "7d") {
		t.Fatalf("expected 7d in filename, got %s", disp)
	}
}
