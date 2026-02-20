package flightapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	baseURL        = "https://serpapi.com/search"
	requestTimeout = 15 * time.Second
	maxRetries     = 3
	baseDelay      = 1 * time.Second
)

// Client communicates with the SerpApi Google Flights API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a SerpApi client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Search queries Google Flights via SerpApi and returns parsed flight results.
// It retries transient failures with exponential backoff.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]FlightResult, error) {
	if params.Currency == "" {
		params.Currency = "USD"
	}
	if params.Adults <= 0 {
		params.Adults = 1
	}
	if params.TravelClass <= 0 {
		params.TravelClass = 1
	}

	u := c.buildSearchURL(params)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			log.Printf("[serpapi] retry %d/%d after %v", attempt, maxRetries, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		results, err := c.doSearch(ctx, u)
		if err == nil {
			return results, nil
		}

		lastErr = err
		if !isRetryable(err) {
			return nil, fmt.Errorf("serpapi search: %w", err)
		}

		log.Printf("[serpapi] attempt %d failed: %v", attempt+1, err)
	}

	return nil, fmt.Errorf("serpapi search: all %d retries exhausted: %w", maxRetries, lastErr)
}

func (c *Client) buildSearchURL(params SearchParams) string {
	q := url.Values{}
	q.Set("engine", "google_flights")
	q.Set("api_key", c.apiKey)
	q.Set("departure_id", params.DepartureID)
	q.Set("arrival_id", params.ArrivalID)
	q.Set("outbound_date", params.OutboundDate.Format("2006-01-02"))
	q.Set("currency", params.Currency)
	q.Set("adults", strconv.Itoa(params.Adults))
	q.Set("travel_class", strconv.Itoa(params.TravelClass))
	q.Set("hl", "en")

	if params.ReturnDate != nil {
		q.Set("return_date", params.ReturnDate.Format("2006-01-02"))
		q.Set("type", "1") // round trip
	} else {
		q.Set("type", "2") // one way
	}

	if params.Stops > 0 {
		q.Set("stops", strconv.Itoa(params.Stops))
	}
	if params.MaxPrice > 0 {
		q.Set("max_price", strconv.Itoa(params.MaxPrice))
	}

	return baseURL + "?" + q.Encode()
}

func (c *Client) doSearch(ctx context.Context, rawURL string) ([]FlightResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &retryableError{err: fmt.Errorf("http request: %w", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &retryableError{err: fmt.Errorf("read response: %w", err)}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &retryableError{err: fmt.Errorf("rate limited (429)")}
	}
	if resp.StatusCode >= 500 {
		return nil, &retryableError{err: fmt.Errorf("server error (%d): %s", resp.StatusCode, truncate(body, 200))}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, truncate(body, 200))
	}

	var serpResp SerpResponse
	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	results := toFlightResults(serpResp.BestFlights)
	results = append(results, toFlightResults(serpResp.OtherFlights)...)

	return results, nil
}

// toFlightResults converts SerpApi flight groups into our internal model.
func toFlightResults(groups []FlightGroup) []FlightResult {
	results := make([]FlightResult, 0, len(groups))
	for _, g := range groups {
		if len(g.Flights) == 0 || g.Price <= 0 {
			continue
		}

		first := g.Flights[0]
		last := g.Flights[len(g.Flights)-1]

		dep, _ := time.Parse("2006-01-02 15:04", first.DepartureAirport.Time)
		arr, _ := time.Parse("2006-01-02 15:04", last.ArrivalAirport.Time)

		airline := first.Airline
		flightNum := first.FlightNumber
		if len(g.Flights) > 1 {
			// Multiple legs — show first airline
			for i := 1; i < len(g.Flights); i++ {
				if g.Flights[i].Airline != airline {
					airline += "," + g.Flights[i].Airline
					break
				}
			}
		}

		results = append(results, FlightResult{
			Price:         float64(g.Price),
			Airline:       airline,
			FlightNumber:  flightNum,
			DepartureCode: first.DepartureAirport.ID,
			ArrivalCode:   last.ArrivalAirport.ID,
			Departure:     dep,
			Arrival:       arr,
			Duration:      g.TotalDuration,
			Stops:         len(g.Layovers),
		})
	}
	return results
}

// retryableError marks an error as safe to retry.
type retryableError struct {
	err error
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

func isRetryable(err error) bool {
	_, ok := err.(*retryableError)
	return ok
}

func truncate(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
