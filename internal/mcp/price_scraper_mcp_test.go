package mcp

import (
	"testing"
)

func BenchmarkScrapePrice(b *testing.B) {
	ps := NewPriceScraper()
	req := PriceRequest{
		ProductName: "tej",
		ShopChain:   "Spar",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ps.ScrapePrice(req)
	}
}

func BenchmarkGetShopCoordinates(b *testing.B) {
	ps := NewPriceScraper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ps.GetShopCoordinates("Spar")
	}
}
