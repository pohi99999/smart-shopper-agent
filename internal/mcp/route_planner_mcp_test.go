package mcp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutePlanner_CalculateRouteMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}

	rp := NewRoutePlanner()
	req := RouteMatrixRequest{
		Source: Coordinates{
			Latitude:  46.8400, // Zalaegerszeg
			Longitude: 16.8439,
		},
		Destinations: map[string]Coordinates{
			"Shop1": {
				Latitude:  46.8450,
				Longitude: 16.8500,
			},
			"Shop2": {
				Latitude:  46.8500,
				Longitude: 16.8400,
			},
		},
	}

	resp, err := rp.CalculateRouteMatrix(req)
	if err != nil {
		t.Logf("Route matrix calculation failed: %v", err)
		return
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(resp))
	}

	for name, route := range resp {
		if route.DistanceKM <= 0 {
			t.Errorf("Expected positive distance for %s, got %f", name, route.DistanceKM)
		}
	}
}

func TestRoutePlanner_CalculateRouteMatrixEmpty(t *testing.T) {
	rp := NewRoutePlanner()
	req := RouteMatrixRequest{
		Source: Coordinates{
			Latitude:  46.8400, // Zalaegerszeg
			Longitude: 16.8439,
		},
		Destinations: map[string]Coordinates{},
	}

	resp, err := rp.CalculateRouteMatrix(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("Expected 0 responses, got %d", len(resp))
	}
}

func TestRoutePlanner_CalculateRouteMatrix_ErrorResponses(t *testing.T) {
	req := RouteMatrixRequest{
		Source: Coordinates{Latitude: 46.84, Longitude: 16.84},
		Destinations: map[string]Coordinates{
			"Shop1": {Latitude: 46.85, Longitude: 16.85},
		},
	}

	t.Run("Non-200 HTTP status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		rp := NewRoutePlanner()
		rp.baseURL = server.URL

		_, err := rp.CalculateRouteMatrix(req)
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

		_, err := rp.CalculateRouteMatrix(req)
		if err == nil {
			t.Fatal("Expected connection error, got nil")
		}
		if !strings.Contains(err.Error(), "OSRM API timeout or connection error") {
			t.Errorf("Expected error to contain 'OSRM API timeout or connection error', got %q", err.Error())
		}
	})

	t.Run("Timeout error", func(t *testing.T) {
		rp := NewRoutePlanner()
		rp.client.Transport = &errorTransport{}

		_, err := rp.CalculateRouteMatrix(req)
		if err == nil {
			t.Fatal("Expected timeout error, got nil")
		}

		expectedErrMsgPrefix := "OSRM API timeout or connection error"
		if !strings.HasPrefix(err.Error(), expectedErrMsgPrefix) {
			t.Errorf("Expected error to start with %q, got %q", expectedErrMsgPrefix, err.Error())
		}
	})
}

type errorTransport struct{}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("mock timeout error")
}

func TestRoutePlanner_CalculateRouteMatrix_TooManyDestinations(t *testing.T) {
	rp := NewRoutePlanner()
	req := RouteMatrixRequest{
		Source:       Coordinates{Latitude: 46.8400, Longitude: 16.8439},
		Destinations: make(map[string]Coordinates),
	}

	for i := 0; i < MaxDestinations+1; i++ {
		req.Destinations[fmt.Sprintf("Shop%d", i)] = Coordinates{
			Latitude:  46.8400 + float64(i)*0.01,
			Longitude: 16.8439 + float64(i)*0.01,
		}
	}

	_, err := rp.CalculateRouteMatrix(req)
	if err == nil {
		t.Fatal("Expected error for too many destinations, got nil")
	}

	expectedErrMsgPrefix := "too many destinations:"
	if !strings.HasPrefix(err.Error(), expectedErrMsgPrefix) {
		t.Errorf("Expected error to start with %q, got %q", expectedErrMsgPrefix, err.Error())
	}
}
