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

func TestScrapePrice(t *testing.T) {
	ps := &PriceScraper{
		shops: map[string]ShopData{
			"Spar": {
				Coordinates: Coordinates{Latitude: 47.0, Longitude: 19.0},
				Prices: map[string]float64{
					"tej": 349.0,
				},
			},
		},
	}

	tests := []struct {
		name     string
		req      PriceRequest
		expected PriceResponse
	}{
		{
			name: "Product and shop exist",
			req: PriceRequest{
				ProductName: "tej",
				ShopChain:   "Spar",
			},
			expected: PriceResponse{
				ProductName: "tej",
				ShopChain:   "Spar",
				Price:       349.0,
				Available:   true,
			},
		},
		{
			name: "Shop exists but product does not",
			req: PriceRequest{
				ProductName: "kenyer",
				ShopChain:   "Spar",
			},
			expected: PriceResponse{
				ProductName: "kenyer",
				ShopChain:   "Spar",
				Price:       299.0,
				Available:   true,
			},
		},
		{
			name: "Shop does not exist",
			req: PriceRequest{
				ProductName: "tej",
				ShopChain:   "Tesco",
			},
			expected: PriceResponse{
				ProductName: "tej",
				ShopChain:   "Tesco",
				Price:       299.0,
				Available:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ps.ScrapePrice(tt.req)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if res != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, res)
			}
		})
	}
}
