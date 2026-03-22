package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jose/flight-scanner/internal/middleware"
	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

var iataRegex = regexp.MustCompile(`^[A-Z]{3}$`)

// RouteHandler handles /api/routes endpoints.
type RouteHandler struct {
	repo      RouteRepository
	monitor   RouteMonitor
	priceRepo PriceHistoryRepository
}

// NewRouteHandler creates a RouteHandler.
func NewRouteHandler(repo RouteRepository, mon RouteMonitor, priceRepo PriceHistoryRepository) *RouteHandler {
	return &RouteHandler{repo: repo, monitor: mon, priceRepo: priceRepo}
}

// RegisterRoutes registers route handler endpoints on the given mux.
func (h *RouteHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/routes", h.List)
	mux.HandleFunc("POST /api/routes", h.Create)
	mux.HandleFunc("PUT /api/routes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/routes/{id}", h.Delete)
	mux.HandleFunc("PATCH /api/routes/{id}/pause", h.Pause)
	mux.HandleFunc("PATCH /api/routes/{id}/resume", h.Resume)
}

func (h *RouteHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	if req.Currency == "" {
		req.Currency = "BRL"
	}

	departure, err := time.Parse("2006-01-02", req.DepartureDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "departure_date must be in YYYY-MM-DD format")
		return
	}
	if departure.Before(time.Now().Truncate(24 * time.Hour)) {
		writeError(w, http.StatusBadRequest, "departure_date must be today or in the future")
		return
	}
	if req.ReturnDate != nil {
		ret, err := time.Parse("2006-01-02", *req.ReturnDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "return_date must be in YYYY-MM-DD format")
			return
		}
		if ret.Before(departure) {
			writeError(w, http.StatusBadRequest, "return_date must be on or after departure_date")
			return
		}
	}

	userID := middleware.UserIDFromContext(r.Context())
	route, err := h.repo.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create route")
		return
	}

	// Start monitoring immediately
	h.monitor.StartRoute(*route)

	writeJSON(w, http.StatusCreated, route)
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	routes, err := h.repo.ListAll(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list routes")
		return
	}
	if routes == nil {
		routes = []models.Route{}
	}

	// Enrich with latest prices
	ids := make([]string, len(routes))
	for i, rt := range routes {
		ids[i] = rt.ID
	}

	prices, _ := h.priceRepo.GetLatestPrices(r.Context(), userID, ids)

	enriched := make([]models.RouteWithPrice, len(routes))
	for i, rt := range routes {
		enriched[i] = models.RouteWithPrice{Route: rt}
		if ph, ok := prices[rt.ID]; ok {
			enriched[i].CurrentPrice = &ph.MinPrice
			enriched[i].LastCheckAt = &ph.CheckedAt
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"routes": enriched})
}

func (h *RouteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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

	userID := middleware.UserIDFromContext(r.Context())
	route, err := h.repo.Update(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "route not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to update route")
		}
		return
	}

	// Restart monitor with updated config
	h.monitor.RestartRoute(*route)

	writeJSON(w, http.StatusOK, route)
}

func (h *RouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	h.monitor.StopRoute(id)

	if err := h.repo.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "route not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to delete route")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (h *RouteHandler) Pause(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.repo.SetStatus(r.Context(), userID, id, "paused"); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "route not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to pause route")
		}
		return
	}

	h.monitor.StopRoute(id)

	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (h *RouteHandler) Resume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	if err := h.repo.SetStatus(r.Context(), userID, id, "active"); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "route not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to resume route")
		}
		return
	}

	route, err := h.repo.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "route not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get route")
		}
		return
	}

	h.monitor.StartRoute(*route)

	writeJSON(w, http.StatusOK, map[string]string{"status": "active"})
}
