package handler

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

// HistoryHandler handles /api/routes/:id/history endpoints.
type HistoryHandler struct {
	priceRepo PriceHistoryRepository
}

// NewHistoryHandler creates a HistoryHandler.
func NewHistoryHandler(priceRepo PriceHistoryRepository) *HistoryHandler {
	return &HistoryHandler{priceRepo: priceRepo}
}

// RegisterRoutes registers history handler endpoints on the given mux.
func (h *HistoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/routes/{id}/history", h.GetHistory)
	mux.HandleFunc("GET /api/routes/{id}/history/export", h.Export)
}

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	routeID := r.PathValue("id")
	if routeID == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

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

func (h *HistoryHandler) Export(w http.ResponseWriter, r *http.Request) {
	routeID := r.PathValue("id")
	if routeID == "" {
		writeError(w, http.StatusBadRequest, "missing route id")
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	format := r.URL.Query().Get("format")

	history, err := h.priceRepo.GetByRoute(r.Context(), routeID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	shortID := routeID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	if format == "json" {
		filename := fmt.Sprintf("price_history_%s_%dd.json", shortID, days)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		writeJSON(w, http.StatusOK, history)
		return
	}

	// Default: CSV
	filename := fmt.Sprintf("price_history_%s_%dd.csv", shortID, days)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"checked_at", "min_price", "max_price", "avg_price", "airline"}); err != nil {
		slog.Error("csv write error", "err", err)
		return
	}

	for _, ph := range history {
		if err := writer.Write([]string{
			ph.CheckedAt.Format("2006-01-02T15:04:05Z"),
			fmt.Sprintf("%.2f", ph.MinPrice),
			fmt.Sprintf("%.2f", ph.MaxPrice),
			fmt.Sprintf("%.2f", ph.AvgPrice),
			ph.Airline,
		}); err != nil {
			slog.Error("csv write error", "err", err)
			return
		}
	}
	writer.Flush()
}
