package agents

import (
	"strings"
	"testing"

	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

func TestOptimizer_Optimize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	// Using the real scraper and planner, but with Zalaegerszeg coordinates
	// to ensure it passes the <50km check.
	scraper := mcp.NewPriceScraper()
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {Coordinates: mcp.Coordinates{Latitude: 46.8451, Longitude: 16.8455}},
		"Interspar": {Coordinates: mcp.Coordinates{Latitude: 46.8413, Longitude: 16.8521}},
	})
	planner := mcp.NewRoutePlanner()
	optimizer := NewOptimizer(planner, scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "kenyér", Quantity: 1},
		},
	}

	prices := map[string]float64{
		"Aldi":      500.0,
		"Interspar": 600.0,
	}

	userCoords := mcp.Coordinates{
		Latitude:  46.8400, // Zalaegerszeg
		Longitude: 16.8439,
	}

	plan, err := optimizer.Optimize(list, prices, userCoords)
	// The real RoutePlanner might fail if the internet is down, or OSRM blocks it
	// So we allow err != nil ONLY if it's a timeout/OSRM error
	if err != nil {
		if strings.Contains(err.Error(), "OSRM API timeout") || strings.Contains(err.Error(), "failed to call") {
			t.Logf("OSRM API failed (expected in CI sometimes): %v", err)
			return // Skip the rest of the test
		}
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(plan.Steps) != 1 {
		t.Fatalf("Expected 1 step in plan, got %d", len(plan.Steps))
	}

	if plan.Steps[0].ShopName != "Aldi" {
		t.Errorf("Expected Aldi (cheaper), got %s", plan.Steps[0].ShopName)
	}
}

func TestOptimizer_DistanceLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	scraper := mcp.NewPriceScraper()
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {Coordinates: mcp.Coordinates{Latitude: 46.8451, Longitude: 16.8455}},
		"Interspar": {Coordinates: mcp.Coordinates{Latitude: 46.8413, Longitude: 16.8521}},
	})
	planner := mcp.NewRoutePlanner()
	optimizer := NewOptimizer(planner, scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "kenyér", Quantity: 1},
		},
	}

	prices := map[string]float64{
		"Aldi": 500.0,
	}

	// Coordinates far away (e.g. New York) -> Distance should be > 50km and thus skip
	userCoords := mcp.Coordinates{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}

	_, err := optimizer.Optimize(list, prices, userCoords)

	if err != nil && (strings.Contains(err.Error(), "OSRM API timeout") || strings.Contains(err.Error(), "failed to call")) {
		t.Logf("OSRM API failed (expected in CI sometimes): %v", err)
		return // Skip the rest of the test
	}

	if err == nil {
		t.Fatalf("Expected error 'no shops found within 50 km', got nil")
	}

	if err.Error() != "no shops found within 50 km" {
		t.Errorf("Expected 'no shops found within 50 km', got: %v", err)
	}
}

func TestOptimizer_EmptyPrices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	scraper := mcp.NewPriceScraper()
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {Coordinates: mcp.Coordinates{Latitude: 46.8451, Longitude: 16.8455}},
		"Interspar": {Coordinates: mcp.Coordinates{Latitude: 46.8413, Longitude: 16.8521}},
	})
	planner := mcp.NewRoutePlanner()
	optimizer := NewOptimizer(planner, scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "kenyér", Quantity: 1},
		},
	}

	prices := map[string]float64{}

	userCoords := mcp.Coordinates{
		Latitude:  46.8400, // Zalaegerszeg
		Longitude: 16.8439,
	}

	_, err := optimizer.Optimize(list, prices, userCoords)

	if err == nil {
		t.Fatalf("Expected error 'no shops found within 50 km', got nil")
	}

	if err.Error() != "no shops found within 50 km" {
		t.Errorf("Expected 'no shops found within 50 km', got: %v", err)
	}
}

func TestOptimizer_EmptyItems(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	scraper := mcp.NewPriceScraper()
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {Coordinates: mcp.Coordinates{Latitude: 46.8451, Longitude: 16.8455}},
		"Interspar": {Coordinates: mcp.Coordinates{Latitude: 46.8413, Longitude: 16.8521}},
	})
	planner := mcp.NewRoutePlanner()
	optimizer := NewOptimizer(planner, scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{},
	}

	prices := map[string]float64{
		"Aldi": 500.0,
	}

	userCoords := mcp.Coordinates{
		Latitude:  46.8400, // Zalaegerszeg
		Longitude: 16.8439,
	}

	plan, err := optimizer.Optimize(list, prices, userCoords)

	if err != nil {
		if strings.Contains(err.Error(), "OSRM API timeout") || strings.Contains(err.Error(), "failed to call") {
			t.Logf("OSRM API failed (expected in CI sometimes): %v", err)
			return // Skip the rest of the test
		}
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(plan.Steps) != 1 {
		t.Fatalf("Expected 1 step in plan, got %d", len(plan.Steps))
	}

	if plan.Steps[0].ShopName != "Aldi" {
		t.Errorf("Expected Aldi, got %s", plan.Steps[0].ShopName)
	}

	if len(plan.Steps[0].Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(plan.Steps[0].Items))
	}
}

func TestOptimizer_GetShopCoordinatesError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	scraper := mcp.NewPriceScraper()
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi": {Coordinates: mcp.Coordinates{Latitude: 46.8451, Longitude: 16.8455}},
	})
	planner := mcp.NewRoutePlanner()
	optimizer := NewOptimizer(planner, scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "kenyér", Quantity: 1},
		},
	}

	prices := map[string]float64{
		"UnknownShop": 500.0,
	}

	userCoords := mcp.Coordinates{
		Latitude:  46.8400,
		Longitude: 16.8439,
	}

	_, err := optimizer.Optimize(list, prices, userCoords)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "not found in database") {
		t.Errorf("Expected 'not found in database', got: %v", err)
	}
}
