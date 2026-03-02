package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

type mockAlertRepo struct {
	alerts []models.Alert
	err    error
}

func (m *mockAlertRepo) ListAll(_ context.Context, _ string) ([]models.Alert, error) {
	return m.alerts, m.err
}

func (m *mockAlertRepo) ListByRoute(_ context.Context, _, routeID string) ([]models.Alert, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []models.Alert
	for _, a := range m.alerts {
		if a.RouteID == routeID {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}

func (m *mockAlertRepo) MarkRead(_ context.Context, _, id string) error {
	if m.err != nil {
		return m.err
	}
	for _, a := range m.alerts {
		if a.ID == id {
			return nil
		}
	}
	return repository.ErrNotFound
}

func TestAlertHandler_List(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		repo      *mockAlertRepo
		wantCode  int
		wantCount int
	}{
		{
			name:      "all alerts",
			path:      "/api/alerts",
			repo:      &mockAlertRepo{alerts: []models.Alert{{ID: "a-1", RouteID: "r-1", AlertPrice: 500, TriggeredPrice: 400, TriggeredAt: time.Now()}}},
			wantCode:  http.StatusOK,
			wantCount: 1,
		},
		{
			name:      "empty",
			path:      "/api/alerts",
			repo:      &mockAlertRepo{},
			wantCode:  http.StatusOK,
			wantCount: 0,
		},
		{
			name: "by route",
			path: "/api/alerts?route_id=r-1",
			repo: &mockAlertRepo{alerts: []models.Alert{
				{ID: "a-1", RouteID: "r-1"},
				{ID: "a-2", RouteID: "r-2"},
			}},
			wantCode:  http.StatusOK,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAlertHandler(tt.repo)
			req := withUserCtx(httptest.NewRequest(http.MethodGet, tt.path, nil))
			w := httptest.NewRecorder()
			h.List(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, w.Code)
			}

			var body struct {
				Alerts []models.Alert `json:"alerts"`
				Count  int            `json:"count"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("json unmarshal: %v", err)
			}
			if body.Count != tt.wantCount {
				t.Fatalf("expected count %d, got %d", tt.wantCount, body.Count)
			}
		})
	}
}

func TestAlertHandler_MarkRead(t *testing.T) {
	tests := []struct {
		name     string
		alertID  string
		repo     *mockAlertRepo
		wantCode int
	}{
		{
			name:     "success",
			alertID:  "a-1",
			repo:     &mockAlertRepo{alerts: []models.Alert{{ID: "a-1", RouteID: "r-1"}}},
			wantCode: http.StatusOK,
		},
		{
			name:     "not found",
			alertID:  "a-999",
			repo:     &mockAlertRepo{alerts: []models.Alert{}},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAlertHandler(tt.repo)
			req := withUserCtx(httptest.NewRequest(http.MethodPatch, "/api/alerts/"+tt.alertID+"/mark-read", nil))
			req.SetPathValue("id", tt.alertID)
			w := httptest.NewRecorder()
			h.MarkRead(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
		})
	}
}
