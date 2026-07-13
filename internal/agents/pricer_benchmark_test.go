package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
	"testing"
)

func BenchmarkPricer_GetPrices(b *testing.B) {
	scraper := mcp.NewPriceScraper()
	pricer := NewPricer(scraper)

	var items []models.ShoppingItem
	for i := 0; i < 1000; i++ {
		items = append(items, models.ShoppingItem{Name: "item", Quantity: 1})
	}

	list := models.ShoppingList{
		Items: items,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pricer.GetPrices(list)
	}
}
