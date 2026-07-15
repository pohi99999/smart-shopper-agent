package mcp

import (
	"net/http"
	"net/http/httptest"
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

func TestRoutePlanner_ErrorResponses(t *testing.T) {
	req := RouteRequest{
		Source: Coordinates{Latitude: 46.84, Longitude: 16.84},
		Destination: Coordinates{Latitude: 46.85, Longitude: 16.85},
	}

	t.Run("Non-200 HTTP status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		rp := NewRoutePlanner()
		rp.baseURL = server.URL

		_, err := rp.CalculateRoute(req)
		if err == nil {
			t.Fatal("Expected error for non-200 status code, got nil")
		}

		expectedErrMsg := "OSRM API returned status: 500 Internal Server Error"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error %q, got %q", expectedErrMsg, err.Error())
		}
	})

	t.Run("Connection error", func(t *testing.T) {
		rp := NewRoutePlanner()
		rp.baseURL = "http://127.0.0.1:0"

		_, err := rp.CalculateRoute(req)
		if err == nil {
			t.Fatal("Expected connection error, got nil")
		}
	})
}
