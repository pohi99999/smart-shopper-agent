package mcp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

type RouteCacheKey struct {
	SourceLat, SourceLon float64
	DestLat, DestLon     float64
}

type RouteCacheEntry struct {
	Response  RouteResponse
	ExpiresAt time.Time
}

type RoutePlanner struct {
	client  *http.Client
	baseURL string
	cache   map[RouteCacheKey]RouteCacheEntry
	mu      sync.RWMutex
}

func NewRoutePlanner() *RoutePlanner {
	return &RoutePlanner{
		baseURL: "https://router.project-osrm.org",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[RouteCacheKey]RouteCacheEntry),
	}
}

func (rp *RoutePlanner) CalculateRouteMatrix(req RouteMatrixRequest) (map[string]RouteResponse, error) {
	if len(req.Destinations) == 0 {
		return map[string]RouteResponse{}, nil
	}

	rp.mu.Lock()
	if rp.cache == nil {
		rp.cache = make(map[RouteCacheKey]RouteCacheEntry)
	}
	rp.mu.Unlock()

	results := make(map[string]RouteResponse)
	uncachedDestinations := make(map[string]Coordinates)
	now := time.Now()

	rp.mu.RLock()
	for name, coord := range req.Destinations {
		key := RouteCacheKey{
			SourceLat: req.Source.Latitude,
			SourceLon: req.Source.Longitude,
			DestLat:   coord.Latitude,
			DestLon:   coord.Longitude,
		}

		if entry, exists := rp.cache[key]; exists && now.Before(entry.ExpiresAt) {
			results[name] = entry.Response
		} else {
			uncachedDestinations[name] = coord
		}
	}
	rp.mu.RUnlock()

	if len(uncachedDestinations) == 0 {
		return results, nil
	}

	// Only query for uncached destinations
	req.Destinations = uncachedDestinations

	// Build the coordinate string: {source_lon},{source_lat};{dest1_lon},{dest1_lat};...
	var coordsBuilder strings.Builder
	// Approximate capacity based on ~25 characters per coordinate pair (e.g. "-123.456789,-123.456789;")
	coordsBuilder.Grow((len(req.Destinations) + 1) * 25)

	coordsBuilder.WriteString(strconv.FormatFloat(req.Source.Longitude, 'f', 6, 64))
	coordsBuilder.WriteString(",")
	coordsBuilder.WriteString(strconv.FormatFloat(req.Source.Latitude, 'f', 6, 64))

	// Ensure consistent order of destinations
	shopNames := make([]string, 0, len(req.Destinations))
	for name, coord := range req.Destinations {
		coordsBuilder.WriteString(";")
		coordsBuilder.WriteString(strconv.FormatFloat(coord.Longitude, 'f', 6, 64))
		coordsBuilder.WriteString(",")
		coordsBuilder.WriteString(strconv.FormatFloat(coord.Latitude, 'f', 6, 64))
		shopNames = append(shopNames, name)
	}

	// sources=0 means the first coordinate is the source
	// destinations=1,2,... means the rest are destinations
	var destIndicesBuilder strings.Builder
	destIndicesBuilder.Grow(len(req.Destinations) * 6)
	for i := 1; i <= len(req.Destinations); i++ {
		if i > 1 {
			destIndicesBuilder.WriteString(";")
		}
		destIndicesBuilder.WriteString(strconv.Itoa(i))
	}

	url := fmt.Sprintf("%s/table/v1/driving/%s?sources=0&destinations=%s&annotations=distance,duration",
		rp.baseURL, coordsBuilder.String(), destIndicesBuilder.String())

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

	expiresAt := time.Now().Add(24 * time.Hour)
	rp.mu.Lock()
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

		resp := RouteResponse{
			DistanceKM:  distanceKM,
			DurationMin: durationMin,
		}
		results[shopName] = resp
		rp.cache[RouteCacheKey{
			SourceLat: req.Source.Latitude,
			SourceLon: req.Source.Longitude,
			DestLat:   req.Destinations[shopName].Latitude,
			DestLon:   req.Destinations[shopName].Longitude,
		}] = RouteCacheEntry{Response: resp, ExpiresAt: expiresAt}

		slog.Debug("Route matrix calculated",
			"shop", shopName,
			"distance_km", distanceKM,
			"duration_min", durationMin,
		)
	}
	rp.mu.Unlock()

	return results, nil
}
