package mcp

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkCalculateRouteMatrix_URLBuilding(b *testing.B) {
	req := RouteMatrixRequest{
		Source:       Coordinates{Latitude: 46.8400, Longitude: 16.8439},
		Destinations: make(map[string]Coordinates),
	}
	for i := 0; i < 100; i++ {
		req.Destinations[fmt.Sprintf("Shop%d", i)] = Coordinates{Latitude: 46.8400 + float64(i)*0.01, Longitude: 16.8439 + float64(i)*0.01}
	}

	rp := NewRoutePlanner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// New logic using strings.Builder
		var coordsBuilder strings.Builder
		// Guess size, roughly 20 chars per coord pair
		coordsBuilder.Grow((len(req.Destinations) + 1) * 25)

		fmt.Fprintf(&coordsBuilder, "%f,%f", req.Source.Longitude, req.Source.Latitude)

		shopNames := make([]string, 0, len(req.Destinations))
		for name, coord := range req.Destinations {
			fmt.Fprintf(&coordsBuilder, ";%f,%f", coord.Longitude, coord.Latitude)
			shopNames = append(shopNames, name)
		}

		var destIndices strings.Builder
		destIndices.Grow(len(req.Destinations) * 4)
		for j := 1; j <= len(req.Destinations); j++ {
			if j > 1 {
				destIndices.WriteString(";")
			}
			fmt.Fprintf(&destIndices, "%d", j)
		}

		_ = fmt.Sprintf("%s/table/v1/driving/%s?sources=0&destinations=%s&annotations=distance,duration",
			rp.baseURL, coordsBuilder.String(), destIndices.String())
		_ = shopNames
	}
}
