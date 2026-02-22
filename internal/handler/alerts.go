package handler

import (
	"net/http"
	"strings"

	"github.com/jose/flight-scanner/internal/models"
	"github.com/jose/flight-scanner/internal/repository"
)

// AlertHandler handles /api/alerts endpoints.
type AlertHandler struct {
	repo *repository.AlertRepo
}

// NewAlertHandler creates an AlertHandler.
func NewAlertHandler(repo *repository.AlertRepo) *AlertHandler {
	return &AlertHandler{repo: repo}
}

// ServeHTTP routes alert requests.
func (h *AlertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/alerts")
	path = strings.TrimRight(path, "/")

	switch {
	case path == "" && r.Method == http.MethodGet:
		h.list(w, r)
	case strings.HasSuffix(path, "/mark-read") && r.Method == http.MethodPatch:
		id := extractID(strings.TrimSuffix(path, "/mark-read"))
		h.markRead(w, r, id)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *AlertHandler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *AlertHandler) markRead(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	if err := h.repo.MarkRead(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "alert not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "read", "id": id})
}
