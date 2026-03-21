package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
)

// SearchHandler handles POST /api/search/flights.
type SearchHandler struct {
	client FlightSearcher
}

// NewSearchHandler creates a SearchHandler.
func NewSearchHandler(client FlightSearcher) *SearchHandler {
	return &SearchHandler{client: client}
}

type searchRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`        // optional, YYYY-MM-DD
	ReturnDate  string `json:"return_date"` // optional, YYYY-MM-DD
	Currency    string `json:"currency"`    // optional, default USD
}

// RegisterRoutes registers search handler endpoints on the given mux.
func (h *SearchHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/search/flights", h.Search)
	mux.HandleFunc("GET /api/search/airports", h.Autocomplete)
}

func (h *SearchHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	results, err := h.client.Autocomplete(r.Context(), q)
	if err != nil {
		slog.Error("airport autocomplete failed", "err", err)
		writeError(w, http.StatusBadGateway, "airport autocomplete temporarily unavailable")
		return
	}

	if results == nil {
		results = []flightapi.AutocompleteResult{}
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
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

	outbound := time.Now().AddDate(0, 0, 1)
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD format")
			return
		}
		outbound = parsed
	}

	params := flightapi.SearchParams{
		DepartureID:  req.Origin,
		ArrivalID:    req.Destination,
		OutboundDate: outbound,
		Currency:     req.Currency,
		Adults:       1,
		TravelClass:  1,
	}

	if req.ReturnDate != "" {
		ret, err := time.Parse("2006-01-02", req.ReturnDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "return_date must be YYYY-MM-DD format")
			return
		}
		params.ReturnDate = &ret
	}

	if params.Currency == "" {
		params.Currency = "USD"
	}

	results, err := h.client.Search(r.Context(), params)
	if err != nil {
		slog.Error("flight search failed", "err", err)
		writeError(w, http.StatusBadGateway, "flight search temporarily unavailable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"origin":      req.Origin,
		"destination": req.Destination,
		"date":        outbound.Format("2006-01-02"),
		"currency":    params.Currency,
		"results":     results,
		"count":       len(results),
	})
}
