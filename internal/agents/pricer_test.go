package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
	"testing"
)

func TestPricer_GetPrices(t *testing.T) {
	scraper := mcp.NewPriceScraper()
	// Inject test data directly to avoid os.Chdir
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {},
		"Interspar": {},
	})

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

func TestPricer_GetPrices_EmptyList(t *testing.T) {
	scraper := mcp.NewPriceScraper()
	// Inject test data directly
	scraper.SetShopsForTesting(map[string]mcp.ShopData{
		"Aldi":      {},
		"Interspar": {},
	})

	pricer := NewPricer(scraper)

	list := models.ShoppingList{
		Items: []models.ShoppingItem{}, // Empty list
	}

	totals, err := pricer.GetPrices(list)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(totals) != 2 {
		t.Errorf("Expected 2 chains (Aldi, Interspar), got %d", len(totals))
	}

	if val, ok := totals["Aldi"]; !ok || val != 0 {
		t.Errorf("Expected Aldi total to be 0, got %f", val)
	}

	if val, ok := totals["Interspar"]; !ok || val != 0 {
		t.Errorf("Expected Interspar total to be 0, got %f", val)
	}
}
