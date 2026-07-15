package agents

import (
	"os"
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
	"testing"
)

func BenchmarkOptimizer(b *testing.B) {
	os.Chdir("../..")
	defer os.Chdir("internal/agents")
	scraper := mcp.NewPriceScraper()
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
		"Tesco":     550.0,
		"Lidl":      480.0,
		"Auchan":    520.0,
	}

	userCoords := mcp.Coordinates{
		Latitude:  46.8400,
		Longitude: 16.8439,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.Optimize(list, prices, userCoords)
	}
}
