package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/monitor"
	"github.com/jose/flight-scanner/internal/repository"
)

var iataRegex = regexp.MustCompile(`^[A-Z]{3}$`)

// RouteHandler handles /api/routes endpoints.
type RouteHandler struct {
	repo    *repository.RouteRepo
	monitor *monitor.Monitor
}

// NewRouteHandler creates a RouteHandler.
func NewRouteHandler(repo *repository.RouteRepo, mon *monitor.Monitor) *RouteHandler {
	return &RouteHandler{repo: repo, monitor: mon}
}

// ServeHTTP routes requests based on method and path.
func (h *RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/routes")
	path = strings.TrimRight(path, "/")

	switch {
	case path == "" && r.Method == http.MethodGet:
		h.list(w, r)
	case path == "" && r.Method == http.MethodPost:
		h.create(w, r)
	case r.Method == http.MethodPut:
		h.update(w, r, extractID(path))
	case r.Method == http.MethodDelete:
		h.delete(w, r, extractID(path))
	case strings.HasSuffix(path, "/pause") && r.Method == http.MethodPatch:
		h.pause(w, r, extractID(strings.TrimSuffix(path, "/pause")))
	case strings.HasSuffix(path, "/resume") && r.Method == http.MethodPatch:
		h.resume(w, r, extractID(strings.TrimSuffix(path, "/resume")))
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *RouteHandler) create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Origin = strings.ToUpper(strings.TrimSpace(req.Origin))
	req.Destination = strings.ToUpper(strings.TrimSpace(req.Destination))

	if !iataRegex.MatchString(req.Origin) || !iataRegex.MatchString(req.Destination) {
		writeError(w, http.StatusBadRequest, "origin and destination must be 3-letter IATA codes")
		return
	}
	if req.AlertPrice <= 0 {
		writeError(w, http.StatusBadRequest, "alert_price must be positive")
		return
	}
	if req.CheckFrequencyMinutes < 0 {
		writeError(w, http.StatusBadRequest, "check_frequency_minutes cannot be negative")
		return
	}

	route, err := h.repo.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create route")
		return
	}

	// Start monitoring immediately
	h.monitor.StartRoute(r.Context(), *route)

	writeJSON(w, http.StatusCreated, route)
}

func (h *RouteHandler) list(w http.ResponseWriter, r *http.Request) {
	routes, err := h.repo.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list routes")
		return
	}
	if routes == nil {
		routes = []models.Route{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"routes": routes})
}

func (h *RouteHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	var req models.UpdateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.AlertPrice != nil && *req.AlertPrice <= 0 {
		writeError(w, http.StatusBadRequest, "alert_price must be positive")
		return
	}
	if req.CheckFrequencyMinutes != nil && *req.CheckFrequencyMinutes < 1 {
		writeError(w, http.StatusBadRequest, "check_frequency_minutes must be at least 1")
		return
	}

	route, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update route")
		return
	}
	if route == nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	// Restart monitor with updated config
	h.monitor.RestartRoute(r.Context(), *route)

	writeJSON(w, http.StatusOK, route)
}

func (h *RouteHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	h.monitor.StopRoute(id)

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (h *RouteHandler) pause(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	if err := h.repo.SetStatus(r.Context(), id, "paused"); err != nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	h.monitor.StopRoute(id)

	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (h *RouteHandler) resume(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	if err := h.repo.SetStatus(r.Context(), id, "active"); err != nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	route, err := h.repo.GetByID(r.Context(), id)
	if err != nil || route == nil {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	h.monitor.StartRoute(r.Context(), *route)

	writeJSON(w, http.StatusOK, map[string]string{"status": "active"})
}

// extractID pulls the route ID from a path like "/uuid" or "/uuid/action".
func extractID(path string) string {
	path = strings.TrimPrefix(path, "/")
	if i := strings.Index(path, "/"); i != -1 {
		return path[:i]
	}
	return path
}
