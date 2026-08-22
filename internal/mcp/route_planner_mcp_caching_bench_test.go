package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkCalculateRouteMatrix(b *testing.B) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response for 100 destinations
		distances := make([]float64, 100)
		durations := make([]float64, 100)
		for i := 0; i < 100; i++ {
			distances[i] = float64(1000 + i)
			durations[i] = float64(60 + i)
		}

		resp := OSRMMatrixResponse{
			Code:      "Ok",
			Distances: [][]float64{distances},
			Durations: [][]float64{durations},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	rp := NewRoutePlanner()
	rp.baseURL = server.URL

	req := RouteMatrixRequest{
		Source:       Coordinates{Latitude: 46.8400, Longitude: 16.8439},
		Destinations: make(map[string]Coordinates),
	}
	for i := 0; i < 100; i++ {
		req.Destinations["Shop"+string(rune(i))] = Coordinates{Latitude: 46.8400 + float64(i)*0.01, Longitude: 16.8439 + float64(i)*0.01}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rp.CalculateRouteMatrix(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
