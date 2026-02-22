package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jose/flight-scanner/internal/repository"
)

// HistoryHandler handles /api/routes/:id/history endpoints.
type HistoryHandler struct {
	priceRepo *repository.PriceHistoryRepo
}

// NewHistoryHandler creates a HistoryHandler.
func NewHistoryHandler(priceRepo *repository.PriceHistoryRepo) *HistoryHandler {
	return &HistoryHandler{priceRepo: priceRepo}
}

// ServeHTTP routes history requests. Expected path: /api/routes/{id}/history[/export]
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract route ID from path: /api/routes/{id}/history...
	path := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}
	routeID := parts[0]

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	// Check if export
	isExport := len(parts) >= 3 && parts[2] == "export"
	format := r.URL.Query().Get("format")

	if isExport {
		h.export(w, r, routeID, days, format)
		return
	}

	h.getHistory(w, r, routeID, days)
}

func (h *HistoryHandler) getHistory(w http.ResponseWriter, r *http.Request, routeID string, days int) {
	history, err := h.priceRepo.GetByRoute(r.Context(), routeID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	stats, err := h.priceRepo.GetStats(r.Context(), routeID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"route_id": routeID,
		"days":     days,
		"history":  history,
		"stats":    stats,
		"count":    len(history),
	})
}

func (h *HistoryHandler) export(w http.ResponseWriter, r *http.Request, routeID string, days int, format string) {
	history, err := h.priceRepo.GetByRoute(r.Context(), routeID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	if format == "json" {
		filename := fmt.Sprintf("price_history_%s_%dd.json", routeID[:8], days)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		writeJSON(w, http.StatusOK, history)
		return
	}

	// Default: CSV
	filename := fmt.Sprintf("price_history_%s_%dd.csv", routeID[:8], days)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	writer := csv.NewWriter(w)
	writer.Write([]string{"checked_at", "min_price", "max_price", "avg_price", "airline"})

	for _, ph := range history {
		writer.Write([]string{
			ph.CheckedAt.Format("2006-01-02T15:04:05Z"),
			fmt.Sprintf("%.2f", ph.MinPrice),
			fmt.Sprintf("%.2f", ph.MaxPrice),
			fmt.Sprintf("%.2f", ph.AvgPrice),
			ph.Airline,
		})
	}
	writer.Flush()
}
