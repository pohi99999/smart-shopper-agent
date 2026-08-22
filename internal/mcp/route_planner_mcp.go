package mcp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
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

// MaxDestinations is the maximum number of destinations allowed in a single request to prevent excessively large matrices.
const MaxDestinations = 50

func NewRoutePlanner() *RoutePlanner {
	return &RoutePlanner{
		baseURL: "https://router.project-osrm.org",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func buildCoordinateStrings(req RouteMatrixRequest) (string, string, []string) {
	var coordsBuilder strings.Builder
	coordsBuilder.Grow((len(req.Destinations) + 1) * 25)

	coordsBuilder.WriteString(strconv.FormatFloat(req.Source.Longitude, 'f', 6, 64))
	coordsBuilder.WriteString(",")
	coordsBuilder.WriteString(strconv.FormatFloat(req.Source.Latitude, 'f', 6, 64))

	shopNames := make([]string, 0, len(req.Destinations))
	for name, coord := range req.Destinations {
		coordsBuilder.WriteString(";")
		coordsBuilder.WriteString(strconv.FormatFloat(coord.Longitude, 'f', 6, 64))
		coordsBuilder.WriteString(",")
		coordsBuilder.WriteString(strconv.FormatFloat(coord.Latitude, 'f', 6, 64))
		shopNames = append(shopNames, name)
	}

	var destIndicesBuilder strings.Builder
	destIndicesBuilder.Grow(len(req.Destinations) * 6)
	for i := 1; i <= len(req.Destinations); i++ {
		if i > 1 {
			destIndicesBuilder.WriteString(";")
		}
		destIndicesBuilder.WriteString(strconv.Itoa(i))
	}

	return coordsBuilder.String(), destIndicesBuilder.String(), shopNames
}

func (rp *RoutePlanner) fetchRouteMatrix(url string) (*OSRMMatrixResponse, error) {
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

	return &osrmResp, nil
}

func (rp *RoutePlanner) CalculateRouteMatrix(req RouteMatrixRequest) (map[string]RouteResponse, error) {
	if len(req.Destinations) == 0 {
		return map[string]RouteResponse{}, nil
	}

	if len(req.Destinations) > MaxDestinations {
		return nil, fmt.Errorf("too many destinations: %d exceeds maximum allowed (%d)", len(req.Destinations), MaxDestinations)
	}

	coordsStr, destIndicesStr, shopNames := buildCoordinateStrings(req)

	url := fmt.Sprintf("%s/table/v1/driving/%s?sources=0&destinations=%s&annotations=distance,duration",
		rp.baseURL, coordsStr, destIndicesStr)

	osrmResp, err := rp.fetchRouteMatrix(url)
	if err != nil {
		return nil, err
	}

	if len(osrmResp.Distances) == 0 || len(osrmResp.Distances[0]) != len(shopNames) {
		return nil, fmt.Errorf("unexpected matrix distances format or length")
	}

	if len(osrmResp.Durations) == 0 || len(osrmResp.Durations[0]) != len(shopNames) {
		return nil, fmt.Errorf("unexpected matrix durations format or length")
	}

	results := make(map[string]RouteResponse)
	for i, shopName := range shopNames {
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
