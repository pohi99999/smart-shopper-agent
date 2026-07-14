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

type RoutePlanner struct {
	client *http.Client
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
