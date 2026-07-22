package mcp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteRequest struct {
	Source      Coordinates `json:"source"`
	Destination Coordinates `json:"destination"`
}

type RouteResponse struct {
	DistanceKM  float64 `json:"distance_km"`
	DurationMin float64 `json:"duration_min"`
}

type RouteMatrixRequest struct {
	Source       Coordinates
	Destinations map[string]Coordinates
}

type OSRMMatrixResponse struct {
	Code      string      `json:"code"`
	Distances [][]float64 `json:"distances"`
	Durations [][]float64 `json:"durations"`
}

type RoutePlanner struct {
	client  *http.Client
	baseURL string
}

func NewRoutePlanner() *RoutePlanner {
	return &RoutePlanner{
		baseURL: "https://router.project-osrm.org",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type OSRMResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

func (rp *RoutePlanner) CalculateRoute(req RouteRequest) (RouteResponse, error) {
	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=false",
		rp.baseURL,
		req.Source.Longitude, req.Source.Latitude,
		req.Destination.Longitude, req.Destination.Latitude)

	resp, err := rp.client.Get(url)
	if err != nil {
		return RouteResponse{}, fmt.Errorf("OSRM API timeout or connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RouteResponse{}, fmt.Errorf("OSRM API returned status: %s", resp.Status)
	}

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return RouteResponse{}, fmt.Errorf("failed to decode OSRM response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return RouteResponse{}, fmt.Errorf("no routes found or OSRM error code: %s", osrmResp.Code)
	}

	route := osrmResp.Routes[0]
	// distance is in meters -> convert to km
	distanceKM := route.Distance / 1000.0
	// duration is in seconds -> convert to minutes
	durationMin := route.Duration / 60.0

	slog.Info("Route calculated",
		"source_lat", req.Source.Latitude,
		"source_lon", req.Source.Longitude,
		"dest_lat", req.Destination.Latitude,
		"dest_lon", req.Destination.Longitude,
		"distance_km", distanceKM,
		"duration_min", durationMin,
	)

	return RouteResponse{
		DistanceKM:  distanceKM,
		DurationMin: durationMin,
	}, nil
}

func (rp *RoutePlanner) CalculateRouteMatrix(req RouteMatrixRequest) (map[string]RouteResponse, error) {
	if len(req.Destinations) == 0 {
		return map[string]RouteResponse{}, nil
	}

	// Build the coordinate string: {source_lon},{source_lat};{dest1_lon},{dest1_lat};...
	coordsStr := fmt.Sprintf("%f,%f", req.Source.Longitude, req.Source.Latitude)

	// Ensure consistent order of destinations
	shopNames := make([]string, 0, len(req.Destinations))
	for name, coord := range req.Destinations {
		coordsStr += fmt.Sprintf(";%f,%f", coord.Longitude, coord.Latitude)
		shopNames = append(shopNames, name)
	}

	// sources=0 means the first coordinate is the source
	// destinations=1,2,... means the rest are destinations
	destIndices := ""
	for i := 1; i <= len(req.Destinations); i++ {
		if i > 1 {
			destIndices += ";"
		}
		destIndices += fmt.Sprintf("%d", i)
	}

	url := fmt.Sprintf("%s/table/v1/driving/%s?sources=0&destinations=%s&annotations=distance,duration",
		rp.baseURL, coordsStr, destIndices)

	resp, err := rp.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("OSRM API timeout or connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM API returned status: %s", resp.Status)
	}

	var osrmResp OSRMMatrixResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return nil, fmt.Errorf("failed to decode OSRM matrix response: %w", err)
	}

	if osrmResp.Code != "Ok" {
		return nil, fmt.Errorf("OSRM matrix error code: %s", osrmResp.Code)
	}

	if len(osrmResp.Distances) == 0 || len(osrmResp.Distances[0]) != len(shopNames) {
		return nil, fmt.Errorf("unexpected matrix distances format or length")
	}

	if len(osrmResp.Durations) == 0 || len(osrmResp.Durations[0]) != len(shopNames) {
		return nil, fmt.Errorf("unexpected matrix durations format or length")
	}

	results := make(map[string]RouteResponse)
	for i, shopName := range shopNames {
		// Index 0 in the response arrays corresponds to the source itself,
		// but since we used destinations=1,2,3... the returned array only contains the requested destinations.
		// Wait, let's verify OSRM response when using sources=0 and destinations=1;2

		// Based on the curl output:
		// "distances": [ [ 1888, 3800.9 ] ]
		// So distances[0] contains exactly the distances to the requested destinations.
		// Therefore index `i` maps exactly to `shopNames[i]`.

		distanceKM := osrmResp.Distances[0][i] / 1000.0
		durationMin := osrmResp.Durations[0][i] / 60.0

		results[shopName] = RouteResponse{
			DistanceKM:  distanceKM,
			DurationMin: durationMin,
		}

		slog.Debug("Route matrix calculated",
			"shop", shopName,
			"distance_km", distanceKM,
			"duration_min", durationMin,
		)
	}

	return results, nil
}
