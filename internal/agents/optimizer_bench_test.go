package agents

import (
	"smart-shopper-agent/internal/mcp"
	"testing"
)

func BenchmarkGatherDestinations_Iterative(b *testing.B) {
	scraper := mcp.NewPriceScraper()
	shops := make(map[string]mcp.ShopData)
	prices := make(map[string]float64)
	for i := 0; i < 100; i++ {
		name := "Shop" + string(rune(i))
		shops[name] = mcp.ShopData{Coordinates: mcp.Coordinates{Latitude: float64(i), Longitude: float64(i)}}
		prices[name] = float64(i)
	}
	scraper.SetShopsForTesting(shops)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destinations := make(map[string]mcp.Coordinates, len(prices))
		for shopName := range prices {
			coords, err := scraper.GetShopCoordinates(shopName)
			if err != nil {
				b.Fatal(err)
			}
			destinations[shopName] = coords
		}
	}
}

func BenchmarkGatherDestinations_Bulk(b *testing.B) {
	scraper := mcp.NewPriceScraper()
	shops := make(map[string]mcp.ShopData)
	prices := make(map[string]float64)
	for i := 0; i < 100; i++ {
		name := "Shop" + string(rune(i))
		shops[name] = mcp.ShopData{Coordinates: mcp.Coordinates{Latitude: float64(i), Longitude: float64(i)}}
		prices[name] = float64(i)
	}
	scraper.SetShopsForTesting(shops)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shopNames := make([]string, 0, len(prices))
		for shopName := range prices {
			shopNames = append(shopNames, shopName)
		}

		_, err := scraper.GetShopCoordinatesBulk(shopNames)
		if err != nil {
			b.Fatal(err)
		}
	}
}
