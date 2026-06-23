package mcp

import (
	"testing"
)

func TestRoutePlanner_CalculateRoute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}

	rp := NewRoutePlanner()
	req := RouteRequest{
		Source: Coordinates{
			Latitude:  46.8400, // Zalaegerszeg
			Longitude: 16.8439,
		},
		Destination: Coordinates{
			Latitude:  46.8450,
			Longitude: 16.8500,
		},
	}

	resp, err := rp.CalculateRoute(req)
	if err != nil {
		t.Logf("Route calculation failed: %v", err)
		return
	}

	if resp.DistanceKM <= 0 {
		t.Errorf("Expected positive distance, got %f", resp.DistanceKM)
	}
}
