package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

// --- mock types ---

type mockRouteRepo struct {
	routes  []models.Route
	created *models.Route
	updated *models.Route
	err     error
}

func (m *mockRouteRepo) Create(_ context.Context, req models.CreateRouteRequest) (*models.Route, error) {
	if m.err != nil {
		return nil, m.err
	}
	r := &models.Route{ID: "r-1", Origin: req.Origin, Destination: req.Destination, AlertPrice: req.AlertPrice, CheckFrequencyMinutes: req.CheckFrequencyMinutes, Status: "active"}
	m.created = r
	return r, nil
}

func (m *mockRouteRepo) ListAll(_ context.Context) ([]models.Route, error) {
	return m.routes, m.err
}

func (m *mockRouteRepo) Update(_ context.Context, id string, req models.UpdateRouteRequest) (*models.Route, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.updated, nil
}

func (m *mockRouteRepo) Delete(_ context.Context, id string) error { return m.err }
func (m *mockRouteRepo) SetStatus(_ context.Context, id, status string) error {
	return m.err
}
func (m *mockRouteRepo) GetByID(_ context.Context, id string) (*models.Route, error) {
	if m.err != nil {
		return nil, m.err
	}
	if len(m.routes) > 0 {
		return &m.routes[0], nil
	}
	return nil, nil
}

type mockMonitor struct {
	started   bool
	stopped   bool
	restarted bool
}

func (m *mockMonitor) StartRoute(_ models.Route) { m.started = true }
func (m *mockMonitor) StopRoute(_ string)         { m.stopped = true }
func (m *mockMonitor) RestartRoute(_ models.Route) {
	m.restarted = true
}

// --- tests ---

func TestRouteHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		repo       *mockRouteRepo
		wantCode   int
		wantRoutes int
	}{
		{
			name:       "empty list",
			repo:       &mockRouteRepo{},
			wantCode:   http.StatusOK,
			wantRoutes: 0,
		},
		{
			name:       "with data",
			repo:       &mockRouteRepo{routes: []models.Route{{ID: "r-1", Origin: "GIG", Destination: "SCL"}}},
			wantCode:   http.StatusOK,
			wantRoutes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewRouteHandler(tt.repo, &mockMonitor{}, &mockPriceHistoryRepo{})
			req := httptest.NewRequest(http.MethodGet, "/api/routes", nil)
			w := httptest.NewRecorder()
			h.List(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, w.Code)
			}

			var body struct {
				Routes []models.Route `json:"routes"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("json unmarshal: %v", err)
			}
			if len(body.Routes) != tt.wantRoutes {
				t.Fatalf("expected %d routes, got %d", tt.wantRoutes, len(body.Routes))
			}
		})
	}
}

func TestRouteHandler_Create(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantCode int
	}{
		{
			name:     "valid",
			payload:  `{"origin":"GIG","destination":"SCL","alert_price":500,"check_frequency_minutes":60}`,
			wantCode: http.StatusCreated,
		},
		{
			name:     "invalid IATA",
			payload:  `{"origin":"XX","destination":"SCL","alert_price":500}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing price",
			payload:  `{"origin":"GIG","destination":"SCL"}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mon := &mockMonitor{}
			h := NewRouteHandler(&mockRouteRepo{}, mon, &mockPriceHistoryRepo{})
			req := httptest.NewRequest(http.MethodPost, "/api/routes", bytes.NewBufferString(tt.payload))
			w := httptest.NewRecorder()
			h.Create(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
			if tt.wantCode == http.StatusCreated && !mon.started {
				t.Fatal("expected monitor to be started")
			}
		})
	}
}

func TestRouteHandler_Update(t *testing.T) {
	price := 400.0
	tests := []struct {
		name     string
		repo     *mockRouteRepo
		wantCode int
	}{
		{
			name:     "success",
			repo:     &mockRouteRepo{updated: &models.Route{ID: "r-1", AlertPrice: price}},
			wantCode: http.StatusOK,
		},
		{
			name:     "not found",
			repo:     &mockRouteRepo{err: repository.ErrNotFound},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mon := &mockMonitor{}
			h := NewRouteHandler(tt.repo, mon, &mockPriceHistoryRepo{})
			payload := `{"alert_price":400}`
			req := httptest.NewRequest(http.MethodPut, "/api/routes/r-1", bytes.NewBufferString(payload))
			req.SetPathValue("id", "r-1")
			w := httptest.NewRecorder()
			h.Update(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
			if tt.wantCode == http.StatusOK && !mon.restarted {
				t.Fatal("expected monitor to be restarted")
			}
		})
	}
}

func TestRouteHandler_Delete(t *testing.T) {
	mon := &mockMonitor{}
	h := NewRouteHandler(&mockRouteRepo{}, mon, &mockPriceHistoryRepo{})

	req := httptest.NewRequest(http.MethodDelete, "/api/routes/r-1", nil)
	req.SetPathValue("id", "r-1")
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !mon.stopped {
		t.Fatal("expected monitor to be stopped")
	}
}

func TestRouteHandler_PauseResume(t *testing.T) {
	tests := []struct {
		name      string
		handler   func(h *RouteHandler) http.HandlerFunc
		repo      *mockRouteRepo
		wantCode  int
		wantStart bool
		wantStop  bool
	}{
		{
			name:     "pause",
			handler:  func(h *RouteHandler) http.HandlerFunc { return h.Pause },
			repo:     &mockRouteRepo{},
			wantCode: http.StatusOK,
			wantStop: true,
		},
		{
			name:      "resume",
			handler:   func(h *RouteHandler) http.HandlerFunc { return h.Resume },
			repo:      &mockRouteRepo{routes: []models.Route{{ID: "r-1", Origin: "GIG", Destination: "SCL", Status: "active"}}},
			wantCode:  http.StatusOK,
			wantStart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mon := &mockMonitor{}
			h := NewRouteHandler(tt.repo, mon, &mockPriceHistoryRepo{})
			req := httptest.NewRequest(http.MethodPatch, "/api/routes/r-1/action", nil)
			req.SetPathValue("id", "r-1")
			w := httptest.NewRecorder()
			tt.handler(h)(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
			if tt.wantStart && !mon.started {
				t.Fatal("expected monitor to be started")
			}
			if tt.wantStop && !mon.stopped {
				t.Fatal("expected monitor to be stopped")
			}
		})
	}
}
