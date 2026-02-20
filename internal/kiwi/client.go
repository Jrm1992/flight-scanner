package kiwi

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
	"strings"
	"time"
)

const (
	baseURL        = "https://tequila-api.kiwi.com"
	searchPath     = "/v2/search"
	requestTimeout = 10 * time.Second
	maxRetries     = 3
	baseDelay      = 500 * time.Millisecond
)

// Client communicates with the Kiwi Tequila flight search API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a Kiwi API client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Search queries the Kiwi /v2/search endpoint and returns parsed flight results.
// It retries transient failures with exponential backoff.
func (c *Client) Search(ctx context.Context, params SearchParams) ([]FlightResult, error) {
	if params.Currency == "" {
		params.Currency = "USD"
	}

	u, err := c.buildSearchURL(params)
	if err != nil {
		return nil, fmt.Errorf("build search URL: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			log.Printf("[kiwi] retry %d/%d after %v", attempt, maxRetries, delay)

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
			return nil, fmt.Errorf("kiwi search: %w", err)
		}

		log.Printf("[kiwi] attempt %d failed: %v", attempt+1, err)
	}

	return nil, fmt.Errorf("kiwi search: all %d retries exhausted: %w", maxRetries, lastErr)
}

func (c *Client) buildSearchURL(params SearchParams) (string, error) {
	q := url.Values{}
	q.Set("fly_from", params.FlyFrom)
	q.Set("fly_to", params.FlyTo)
	q.Set("date_from", params.DateFrom.Format("02/01/2006"))
	q.Set("date_to", params.DateTo.Format("02/01/2006"))
	q.Set("curr", params.Currency)

	if params.MaxStopovers > 0 {
		q.Set("max_stopovers", strconv.Itoa(params.MaxStopovers))
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	return baseURL + searchPath + "?" + q.Encode(), nil
}

func (c *Client) doSearch(ctx context.Context, rawURL string) ([]FlightResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("apikey", c.apiKey)

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

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return toFlightResults(searchResp.Data), nil
}

// toFlightResults converts raw API flights into our internal model.
func toFlightResults(flights []Flight) []FlightResult {
	results := make([]FlightResult, 0, len(flights))
	for _, f := range flights {
		airline := ""
		if len(f.Airlines) > 0 {
			airline = strings.Join(f.Airlines, ",")
		}

		dep, _ := time.Parse("2006-01-02T15:04:05.000Z", f.LocalDeparture)
		arr, _ := time.Parse("2006-01-02T15:04:05.000Z", f.LocalArrival)

		results = append(results, FlightResult{
			Price:     f.Price,
			Airline:   airline,
			FlyFrom:   f.FlyFrom,
			FlyTo:     f.FlyTo,
			Departure: dep,
			Arrival:   arr,
			DeepLink:  f.DeepLink,
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
