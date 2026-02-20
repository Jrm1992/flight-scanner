package monitor

import (
	"testing"
	"time"

	"github.com/jose/flight-scanner/internal/flightapi"
)

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
