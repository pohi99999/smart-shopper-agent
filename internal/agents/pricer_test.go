package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
	"testing"
)

import "os"

func TestPricer_GetPrices(t *testing.T) {
	// Change to root dir so internal/data/prices.json can be found
	os.Chdir("../..")
	defer os.Chdir("internal/agents")
	// mcp.PriceScraper uses the real internal/data/prices.json if available,
	// or falls back to a 299.0 default. Let's test the fallback/general behavior
	// using the actual scraper struct (as it doesn't do real external network calls
	// currently, just local file reads).

	scraper := mcp.NewPriceScraper()
	pricer := NewPricer(scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{
			{Name: "kenyér", Quantity: 2},
			{Name: "tej", Quantity: 1},
		},
	}

	totals, err := pricer.GetPrices(list)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(totals) != 2 {
		t.Errorf("Expected 2 chains (Aldi, Interspar), got %d", len(totals))
	}

	if _, ok := totals["Aldi"]; !ok {
		t.Errorf("Expected Aldi in totals")
	}

	if _, ok := totals["Interspar"]; !ok {
		t.Errorf("Expected Interspar in totals")
	}

	// Since fallback price is 299.0, or real prices are parsed, total should be > 0
	if totals["Aldi"] <= 0 {
		t.Errorf("Expected Aldi total > 0, got %f", totals["Aldi"])
	}
}
