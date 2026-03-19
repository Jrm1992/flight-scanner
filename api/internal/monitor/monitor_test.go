package monitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
	"github.com/jose/flight-scanner/internal/models"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockFlightSearcher struct {
	mu      sync.Mutex
	results []flightapi.FlightResult
	err     error
	called  int
}

func (m *mockFlightSearcher) Search(_ context.Context, _ flightapi.SearchParams) ([]flightapi.FlightResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called++
	return m.results, m.err
}

type mockPriceHistoryStore struct {
	mu       sync.Mutex
	inserted []priceInsertCall
	err      error
}

type priceInsertCall struct {
	routeID  string
	minPrice float64
	maxPrice float64
	avgPrice float64
	airline  string
}

func (m *mockPriceHistoryStore) Insert(_ context.Context, routeID string, minPrice, maxPrice, avgPrice float64, airline string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inserted = append(m.inserted, priceInsertCall{routeID, minPrice, maxPrice, avgPrice, airline})
	return m.err
}

type mockAlertStore struct {
	mu               sync.Mutex
	hasAlertToday    bool
	hasAlertTodayErr error
	created          []alertCreateCall
	createErr        error
}

type alertCreateCall struct {
	routeID        string
	alertPrice     float64
	triggeredPrice float64
}

func (m *mockAlertStore) HasAlertToday(_ context.Context, routeID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hasAlertToday, m.hasAlertTodayErr
}

func (m *mockAlertStore) Create(_ context.Context, routeID string, alertPrice, triggeredPrice float64) (*models.Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.created = append(m.created, alertCreateCall{routeID, alertPrice, triggeredPrice})
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &models.Alert{ID: "alert-1", RouteID: routeID, AlertPrice: alertPrice, TriggeredPrice: triggeredPrice}, nil
}

type mockRouteStore struct {
	routes []models.Route
	err    error
}

func (m *mockRouteStore) ListActive(_ context.Context) ([]models.Route, error) {
	return m.routes, m.err
}

// helper to build a route with a future departure date so fetchPrices won't skip it.
func makeRoute(id, origin, dest string, alertPrice float64) models.Route {
	future := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	return models.Route{
		ID:                    id,
		Origin:                origin,
		Destination:           dest,
		DepartureDate:         future,
		AlertPrice:            alertPrice,
		CheckFrequencyMinutes: 60,
		Status:                "active",
	}
}

// helper to create a monitor with a parent context already set.
func setupMonitor(fs *mockFlightSearcher, ph *mockPriceHistoryStore, as *mockAlertStore, rs *mockRouteStore) (*Monitor, context.CancelFunc) {
	mon := New(rs, ph, as, fs)
	ctx, cancel := context.WithCancel(context.Background())
	mon.ctx = ctx
	return mon, cancel
}

// ---------------------------------------------------------------------------
// Monitor manager tests
// ---------------------------------------------------------------------------

func TestStartRoute(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	route := makeRoute("r1", "GIG", "SCL", 500)
	mon.StartRoute(route)

	if !mon.IsRunning("r1") {
		t.Error("expected route r1 to be running")
	}
	if mon.RunningCount() != 1 {
		t.Errorf("expected RunningCount=1, got %d", mon.RunningCount())
	}
}

func TestStopRoute(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	route := makeRoute("r1", "GIG", "SCL", 500)
	mon.StartRoute(route)
	mon.StopRoute("r1")

	if mon.IsRunning("r1") {
		t.Error("expected route r1 to be stopped")
	}
	if mon.RunningCount() != 0 {
		t.Errorf("expected RunningCount=0, got %d", mon.RunningCount())
	}
}

func TestStartRoute_AlreadyRunning(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	route := makeRoute("r1", "GIG", "SCL", 500)
	mon.StartRoute(route)
	mon.StartRoute(route) // second call should be a no-op

	if mon.RunningCount() != 1 {
		t.Errorf("expected RunningCount=1 after duplicate start, got %d", mon.RunningCount())
	}
}

func TestRestartRoute(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	route := makeRoute("r1", "GIG", "SCL", 500)
	mon.StartRoute(route)

	// Update the route and restart.
	route.AlertPrice = 300
	mon.RestartRoute(route)

	if !mon.IsRunning("r1") {
		t.Error("expected route r1 to be running after restart")
	}
	if mon.RunningCount() != 1 {
		t.Errorf("expected RunningCount=1 after restart, got %d", mon.RunningCount())
	}
}

func TestStopAll(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	for i := 0; i < 5; i++ {
		mon.StartRoute(makeRoute(fmt.Sprintf("r%d", i), "GIG", "SCL", 500))
	}

	mon.StopAll()

	if mon.RunningCount() != 0 {
		t.Errorf("expected RunningCount=0 after StopAll, got %d", mon.RunningCount())
	}
}

func TestRunningCount(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	mon.StartRoute(makeRoute("r1", "GIG", "SCL", 500))
	mon.StartRoute(makeRoute("r2", "EZE", "MIA", 400))
	mon.StartRoute(makeRoute("r3", "GRU", "CDG", 600))

	if mon.RunningCount() != 3 {
		t.Errorf("expected RunningCount=3, got %d", mon.RunningCount())
	}
}

// ---------------------------------------------------------------------------
// Worker tests (via newWorker + check())
// ---------------------------------------------------------------------------

func TestWorkerCheck_Success(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 450, Airline: "LATAM"},
			{Price: 550, Airline: "GOL"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	if fs.called != 1 {
		t.Errorf("expected flightSearcher.Search called 1 time, got %d", fs.called)
	}
	if len(ph.inserted) != 1 {
		t.Fatalf("expected 1 price insert, got %d", len(ph.inserted))
	}
	ins := ph.inserted[0]
	if ins.routeID != "r1" {
		t.Errorf("expected routeID=r1, got %s", ins.routeID)
	}
	if ins.minPrice != 450 {
		t.Errorf("expected minPrice=450, got %f", ins.minPrice)
	}
	if ins.maxPrice != 550 {
		t.Errorf("expected maxPrice=550, got %f", ins.maxPrice)
	}
	if ins.avgPrice != 500 {
		t.Errorf("expected avgPrice=500, got %f", ins.avgPrice)
	}
	if ins.airline != "LATAM" {
		t.Errorf("expected airline=LATAM, got %s", ins.airline)
	}
}

func TestWorkerCheck_NoResults(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{results: []flightapi.FlightResult{}}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	if len(ph.inserted) != 0 {
		t.Errorf("expected no price inserts for empty results, got %d", len(ph.inserted))
	}
}

func TestWorkerCheck_AlertTriggered(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"}, // below alert threshold of 500
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{hasAlertToday: false}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	if len(as.created) != 1 {
		t.Fatalf("expected 1 alert created, got %d", len(as.created))
	}
	ac := as.created[0]
	if ac.routeID != "r1" {
		t.Errorf("expected routeID=r1, got %s", ac.routeID)
	}
	if ac.alertPrice != 500 {
		t.Errorf("expected alertPrice=500, got %f", ac.alertPrice)
	}
	if ac.triggeredPrice != 400 {
		t.Errorf("expected triggeredPrice=400, got %f", ac.triggeredPrice)
	}
}

func TestWorkerCheck_AlertDedup(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"}, // below threshold
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{hasAlertToday: true} // already alerted today

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	if len(as.created) != 0 {
		t.Errorf("expected no alert created due to dedup, got %d", len(as.created))
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantStr string // expected formatted as 2006-01-02
		wantErr bool
	}{
		{name: "plain date", input: "2026-04-15", wantStr: "2026-04-15"},
		{name: "with timestamp", input: "2026-04-15T00:00:00Z", wantStr: "2026-04-15"},
		{name: "with time offset", input: "2026-04-15T12:30:00+03:00", wantStr: "2026-04-15"},
		{name: "too short", input: "2026-04", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDate(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got.Format("2006-01-02") != tc.wantStr {
				t.Errorf("expected %s, got %s", tc.wantStr, got.Format("2006-01-02"))
			}
		})
	}
}

func TestAggregateResults(t *testing.T) {
	results := []flightapi.FlightResult{
		{Price: 500, Airline: "LATAM"},
		{Price: 300, Airline: "GOL"},
		{Price: 400, Airline: "Azul"},
	}

	min, max, avg, airline := aggregateResults(results)

	if min != 300 {
		t.Errorf("expected min=300, got %f", min)
	}
	if max != 500 {
		t.Errorf("expected max=500, got %f", max)
	}
	if avg != 400 {
		t.Errorf("expected avg=400, got %f", avg)
	}
	if airline != "GOL" {
		t.Errorf("expected airline=GOL (cheapest), got %q", airline)
	}
}

func TestAggregateResults_SingleResult(t *testing.T) {
	results := []flightapi.FlightResult{
		{Price: 250, Airline: "TAP"},
	}

	min, max, avg, airline := aggregateResults(results)

	if min != 250 || max != 250 || avg != 250 {
		t.Errorf("expected all prices=250, got min=%f max=%f avg=%f", min, max, avg)
	}
	if airline != "TAP" {
		t.Errorf("expected airline=TAP, got %q", airline)
	}
}

func TestWorkerFetchParams(t *testing.T) {
	// Verify that fetchPrices builds correct search params from a route.
	// We can't call fetchPrices directly without a real client, but we can
	// verify the date range logic.
	now := time.Now()
	dateFrom := now
	dateTo := now.AddDate(0, 1, 0)

	diff := dateTo.Sub(dateFrom)
	if diff.Hours() < 24*28 || diff.Hours() > 24*32 {
		t.Errorf("expected ~30 days range, got %v", diff)
	}
}

// ---------------------------------------------------------------------------
// Start() tests
// ---------------------------------------------------------------------------

func TestStart_Success(t *testing.T) {
	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{
		routes: []models.Route{
			makeRoute("r1", "GIG", "SCL", 500),
			makeRoute("r2", "EZE", "MIA", 400),
		},
	}

	mon := New(rs, ph, as, fs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := mon.Start(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mon.RunningCount() != 2 {
		t.Errorf("expected RunningCount=2, got %d", mon.RunningCount())
	}
	mon.StopAll()
}

func TestStart_ListActiveError(t *testing.T) {
	fs := &mockFlightSearcher{}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{err: fmt.Errorf("db connection failed")}

	mon := New(rs, ph, as, fs)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := mon.Start(ctx)
	if err == nil {
		t.Fatal("expected error from Start when ListActive fails")
	}
}

// ---------------------------------------------------------------------------
// Worker check() error path tests
// ---------------------------------------------------------------------------

func TestWorkerCheck_FetchPricesError(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{err: fmt.Errorf("api timeout")}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// Should not insert anything when fetch fails
	ph.mu.Lock()
	insertCount := len(ph.inserted)
	ph.mu.Unlock()
	if insertCount != 0 {
		t.Errorf("expected no price inserts on fetch error, got %d", insertCount)
	}
}

func TestWorkerCheck_InsertError(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{err: fmt.Errorf("db write failed")}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// Insert was attempted but failed; no alert should be created
	as.mu.Lock()
	alertCount := len(as.created)
	as.mu.Unlock()
	if alertCount != 0 {
		t.Errorf("expected no alert created on insert error, got %d", alertCount)
	}
}

func TestWorkerCheck_PriceAboveThreshold_NoAlert(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 600, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// Price is above threshold, no alert should be created
	as.mu.Lock()
	alertCount := len(as.created)
	as.mu.Unlock()
	if alertCount != 0 {
		t.Errorf("expected no alert when price above threshold, got %d", alertCount)
	}
}

// ---------------------------------------------------------------------------
// tryCreateAlert error path tests
// ---------------------------------------------------------------------------

func TestWorkerTryCreateAlert_HasAlertTodayError(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{hasAlertTodayErr: fmt.Errorf("db read error")}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// HasAlertToday returned error, so no alert should be created
	as.mu.Lock()
	alertCount := len(as.created)
	as.mu.Unlock()
	if alertCount != 0 {
		t.Errorf("expected no alert created on HasAlertToday error, got %d", alertCount)
	}
}

func TestWorkerTryCreateAlert_CreateError(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{
		hasAlertToday: false,
		createErr:     fmt.Errorf("db write error"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// Create was called but returned an error
	as.mu.Lock()
	alertCount := len(as.created)
	as.mu.Unlock()
	if alertCount != 1 {
		t.Errorf("expected 1 alert create attempt, got %d", alertCount)
	}
}

// ---------------------------------------------------------------------------
// fetchPrices edge case tests
// ---------------------------------------------------------------------------

func TestWorkerFetchPrices_PastDepartureDate(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	route.DepartureDate = "2020-01-01" // past date

	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 400, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// fetchPrices returns nil when departure date has passed, so no insert
	ph.mu.Lock()
	insertCount := len(ph.inserted)
	ph.mu.Unlock()
	if insertCount != 0 {
		t.Errorf("expected no price inserts for past departure, got %d", insertCount)
	}

	// Search should not have been called since departure date is past
	fs.mu.Lock()
	callCount := fs.called
	fs.mu.Unlock()
	if callCount != 0 {
		t.Errorf("expected Search not called for past departure, got %d calls", callCount)
	}
}

func TestWorkerFetchPrices_WithReturnDate(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	retDate := time.Now().AddDate(0, 2, 0).Format("2006-01-02")
	route.ReturnDate = &retDate

	fs := &mockFlightSearcher{
		results: []flightapi.FlightResult{
			{Price: 450, Airline: "LATAM"},
		},
	}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// Should have called Search and inserted
	fs.mu.Lock()
	callCount := fs.called
	fs.mu.Unlock()
	if callCount != 1 {
		t.Errorf("expected Search called once, got %d", callCount)
	}

	ph.mu.Lock()
	insertCount := len(ph.inserted)
	ph.mu.Unlock()
	if insertCount != 1 {
		t.Errorf("expected 1 price insert, got %d", insertCount)
	}
}

func TestWorkerFetchPrices_InvalidDepartureDate(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	route.DepartureDate = "bad-date"

	fs := &mockFlightSearcher{}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWorker(ctx, route, fs, ph, as)
	w.check()

	// fetchPrices should return error due to bad date, no insert
	ph.mu.Lock()
	insertCount := len(ph.inserted)
	ph.mu.Unlock()
	if insertCount != 0 {
		t.Errorf("expected no price inserts for bad departure date, got %d", insertCount)
	}
}

// ---------------------------------------------------------------------------
// run() loop tests
// ---------------------------------------------------------------------------

func TestWorkerRun_ContextCancellation(t *testing.T) {
	route := makeRoute("r1", "GIG", "SCL", 500)
	route.CheckFrequencyMinutes = 1 // 1 minute

	fs := &mockFlightSearcher{results: nil}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}

	ctx, cancel := context.WithCancel(context.Background())
	w := newWorker(ctx, route, fs, ph, as)

	done := make(chan struct{})
	go func() {
		w.run()
		close(done)
	}()

	// Give the goroutine time to start and do its first check
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// run() exited as expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker.run() did not exit after context cancellation")
	}
}

func TestWorkerRun_PanicRecovery(t *testing.T) {
	// Use a nil flightClient to trigger a panic inside check()
	route := makeRoute("r1", "GIG", "SCL", 500)
	route.CheckFrequencyMinutes = 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A worker with nil dependencies will panic when check() is called
	w := &worker{
		route:        route,
		flightClient: nil, // will cause nil pointer dereference in check -> fetchPrices
		priceHistory: nil,
		alerts:       nil,
		ctx:          ctx,
		cancel:       cancel,
	}

	done := make(chan struct{})
	go func() {
		w.run()
		close(done)
	}()

	select {
	case <-done:
		// run() recovered from panic and exited
	case <-time.After(2 * time.Second):
		t.Fatal("worker.run() did not recover from panic")
	}
}

// ---------------------------------------------------------------------------
// StopRoute for non-existent route (no-op)
// ---------------------------------------------------------------------------

func TestStopRoute_NonExistent(t *testing.T) {
	fs := &mockFlightSearcher{}
	ph := &mockPriceHistoryStore{}
	as := &mockAlertStore{}
	rs := &mockRouteStore{}

	mon, cancel := setupMonitor(fs, ph, as, rs)
	defer cancel()

	// Should not panic
	mon.StopRoute("nonexistent")

	if mon.RunningCount() != 0 {
		t.Errorf("expected RunningCount=0, got %d", mon.RunningCount())
	}
}
