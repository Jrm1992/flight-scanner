package handler

import (
	"errors"
	"net/http"

	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

// AlertHandler handles /api/alerts endpoints.
type AlertHandler struct {
	repo AlertRepository
}

// NewAlertHandler creates an AlertHandler.
func NewAlertHandler(repo AlertRepository) *AlertHandler {
	return &AlertHandler{repo: repo}
}

// RegisterRoutes registers alert handler endpoints on the given mux.
func (h *AlertHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/alerts", h.List)
	mux.HandleFunc("PATCH /api/alerts/{id}/mark-read", h.MarkRead)
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	routeID := r.URL.Query().Get("route_id")

	var (
		alerts []models.Alert
		err    error
	)

	if routeID != "" {
		alerts, err = h.repo.ListByRoute(r.Context(), routeID)
	} else {
		alerts, err = h.repo.ListAll(r.Context())
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}
	if alerts == nil {
		alerts = []models.Alert{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"alerts": alerts, "count": len(alerts)})
}

func (h *AlertHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	if err := h.repo.MarkRead(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "alert not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to mark alert read")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "read", "id": id})
}
