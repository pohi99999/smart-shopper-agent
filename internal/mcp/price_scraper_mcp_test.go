package mcp

import (
	"testing"
)

func BenchmarkGetShopCoordinates(b *testing.B) {
	ps := NewPriceScraper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ps.GetShopCoordinates("Spar")
	}
}

func TestGetShopCoordinates(t *testing.T) {
	ps := &PriceScraper{}
	ps.SetShopsForTesting(map[string]ShopData{
		"Aldi": {
			Coordinates: Coordinates{Latitude: 46.8451, Longitude: 16.8455},
		},
	})

	tests := []struct {
		name        string
		shopChain   string
		wantErr     bool
		expectedLat float64
		expectedLon float64
	}{
		{
			name:        "Success - Existing shop",
			shopChain:   "Aldi",
			wantErr:     false,
			expectedLat: 46.8451,
			expectedLon: 16.8455,
		},
		{
			name:      "Error - Missing shop",
			shopChain: "MissingShop",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coords, err := ps.GetShopCoordinates(tt.shopChain)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if coords.Latitude != tt.expectedLat || coords.Longitude != tt.expectedLon {
				t.Errorf("expected coordinates (%f, %f), got (%f, %f)", tt.expectedLat, tt.expectedLon, coords.Latitude, coords.Longitude)
			}
		})
	}
}

func TestScrapePrices(t *testing.T) {
	ps := &PriceScraper{}
	ps.SetShopsForTesting(map[string]ShopData{
		"Spar": {
			Coordinates: Coordinates{Latitude: 47.0, Longitude: 19.0},
			Prices: map[string]float64{
				"tej": 349.0,
				"viz": 150.0,
			},
		},
	})

	tests := []struct {
		name     string
		req      PriceBatchRequest
		expected []PriceResponse
	}{
		{
			name: "All products and shop exist",
			req: PriceBatchRequest{
				ShopChain:    "Spar",
				ProductNames: []string{"tej", "viz"},
			},
			expected: []PriceResponse{
				{
					ProductName: "tej",
					ShopChain:   "Spar",
					Price:       349.0,
					Available:   true,
				},
				{
					ProductName: "viz",
					ShopChain:   "Spar",
					Price:       150.0,
					Available:   true,
				},
			},
		},
		{
			name: "Shop exists but some products do not",
			req: PriceBatchRequest{
				ShopChain:    "Spar",
				ProductNames: []string{"tej", "kenyer"},
			},
			expected: []PriceResponse{
				{
					ProductName: "tej",
					ShopChain:   "Spar",
					Price:       349.0,
					Available:   true,
				},
				{
					ProductName: "kenyer",
					ShopChain:   "Spar",
					Price:       0.0,
					Available:   false,
				},
			},
		},
		{
			name: "Shop does not exist",
			req: PriceBatchRequest{
				ShopChain:    "Tesco",
				ProductNames: []string{"tej", "viz"},
			},
			expected: []PriceResponse{
				{
					ProductName: "tej",
					ShopChain:   "Tesco",
					Price:       0.0,
					Available:   false,
				},
				{
					ProductName: "viz",
					ShopChain:   "Tesco",
					Price:       0.0,
					Available:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ps.ScrapePrices(tt.req)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(res) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(res))
			}

			for i := range res {
				if res[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %+v, got %+v", i, tt.expected[i], res[i])
				}
			}
		})
	}
}

func TestGetShopChains(t *testing.T) {
	tests := []struct {
		name     string
		shops    map[string]ShopData
		expected []string
	}{
		{
			name:     "Empty shops",
			shops:    map[string]ShopData{},
			expected: []string{},
		},
		{
			name: "Single shop",
			shops: map[string]ShopData{
				"Aldi": {},
			},
			expected: []string{"Aldi"},
		},
		{
			name: "Multiple shops",
			shops: map[string]ShopData{
				"Aldi":      {},
				"Interspar": {},
				"Tesco":     {},
			},
			expected: []string{"Aldi", "Interspar", "Tesco"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriceScraper{}
			ps.SetShopsForTesting(tt.shops)

			chains := ps.GetShopChains()
			if len(chains) != len(tt.expected) {
				t.Fatalf("expected %d chains, got %d", len(tt.expected), len(chains))
			}

			found := map[string]bool{}
			for _, c := range chains {
				found[c] = true
			}

			for _, e := range tt.expected {
				if !found[e] {
					t.Errorf("expected to find chain %s", e)
				}
			}
		})
	}
}

func TestGetShopCoordinatesBulk(t *testing.T) {
	ps := &PriceScraper{}
	ps.SetShopsForTesting(map[string]ShopData{
		"Aldi": {
			Coordinates: Coordinates{Latitude: 46.8451, Longitude: 16.8455},
		},
		"Spar": {
			Coordinates: Coordinates{Latitude: 47.0, Longitude: 19.0},
		},
	})

	tests := []struct {
		name       string
		shopChains []string
		wantErr    bool
		expected   map[string]Coordinates
	}{
		{
			name:       "Success - Existing shops",
			shopChains: []string{"Aldi", "Spar"},
			wantErr:    false,
			expected: map[string]Coordinates{
				"Aldi": {Latitude: 46.8451, Longitude: 16.8455},
				"Spar": {Latitude: 47.0, Longitude: 19.0},
			},
		},
		{
			name:       "Success - Empty input",
			shopChains: []string{},
			wantErr:    false,
			expected:   map[string]Coordinates{},
		},
		{
			name:       "Error - Missing shop",
			shopChains: []string{"Aldi", "MissingShop"},
			wantErr:    true,
			expected:   nil,
		},
		{
			name:       "Partial error - one shop exists, another missing",
			shopChains: []string{"Aldi", "MissingShop"},
			wantErr:    true,
			expected:   nil,
		},
		{
			name:       "Partial error - first missing, second exists",
			shopChains: []string{"MissingShop", "Aldi"},
			wantErr:    true,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coords, err := ps.GetShopCoordinatesBulk(tt.shopChains)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
				if coords != nil {
					t.Errorf("expected nil coords on error, got %v", coords)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(coords) != len(tt.expected) {
				t.Fatalf("expected map of length %d, got %d", len(tt.expected), len(coords))
			}

			for k, expectedCoord := range tt.expected {
				if actualCoord, ok := coords[k]; !ok {
					t.Errorf("expected to find coordinates for shop %s", k)
				} else if actualCoord.Latitude != expectedCoord.Latitude || actualCoord.Longitude != expectedCoord.Longitude {
					t.Errorf("expected coordinates (%f, %f) for shop %s, got (%f, %f)", expectedCoord.Latitude, expectedCoord.Longitude, k, actualCoord.Latitude, actualCoord.Longitude)
				}
			}
		})
	}
}

func TestUpdateCachedChains(t *testing.T) {
	tests := []struct {
		name     string
		shops    map[string]ShopData
		expected []string
	}{
		{
			name:     "Empty map",
			shops:    map[string]ShopData{},
			expected: []string{},
		},
		{
			name: "Map with items",
			shops: map[string]ShopData{
				"Spar":  {},
				"Tesco": {},
				"Aldi":  {},
			},
			expected: []string{"Aldi", "Spar", "Tesco"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PriceScraper{
				shops: tt.shops,
			}
			ps.updateCachedChains()

			if len(ps.cachedChains) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(ps.cachedChains))
			}

			found := make(map[string]bool)
			for _, c := range ps.cachedChains {
				found[c] = true
			}

			for _, e := range tt.expected {
				if !found[e] {
					t.Errorf("expected to find chain %s", e)
				}
			}
		})
	}
}

func TestUpdateCachedChains_StateChange(t *testing.T) {
	ps := &PriceScraper{
		shops: map[string]ShopData{
			"Aldi": {},
		},
	}
	ps.updateCachedChains()

	if len(ps.cachedChains) != 1 || ps.cachedChains[0] != "Aldi" {
		t.Fatalf("expected [Aldi], got %v", ps.cachedChains)
	}

	ps.shops = map[string]ShopData{
		"Spar":  {},
		"Tesco": {},
	}
	ps.updateCachedChains()

	if len(ps.cachedChains) != 2 {
		t.Fatalf("expected length 2, got %d", len(ps.cachedChains))
	}

	found := make(map[string]bool)
	for _, c := range ps.cachedChains {
		found[c] = true
	}

	expected := []string{"Spar", "Tesco"}
	for _, e := range expected {
		if !found[e] {
			t.Errorf("expected to find chain %s after state change", e)
		}
	}
}
