package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"smart-shopper-agent/internal/utils"
)

type PriceResponse struct {
	ProductName string  `json:"product_name"`
	ShopChain   string  `json:"shop_chain"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

type PriceBatchRequest struct {
	ShopChain    string   `json:"shop_chain"`
	ProductNames []string `json:"product_names"`
}

type ShopData struct {
	Coordinates Coordinates        `json:"coordinates"`
	Prices      map[string]float64 `json:"prices"`
}

type PriceScraper struct {
	shops        map[string]ShopData
	cachedChains []string
}

func NewPriceScraper() *PriceScraper {
	ps := &PriceScraper{
		shops: make(map[string]ShopData),
	}

	// Olvassuk be a JSON fájlt egyszer, inicializáláskor
	data, err := os.ReadFile(utils.GetPricesFilePath())
	if err == nil {
		// Ha nincs hiba, próbáljuk meg parse-olni
		_ = json.Unmarshal(data, &ps.shops)
	}

	ps.updateCachedChains()
	return ps
}

func (ps *PriceScraper) ScrapePrices(req PriceBatchRequest) ([]PriceResponse, error) {
	responses := make([]PriceResponse, len(req.ProductNames))
	shopData, shopExists := ps.shops[req.ShopChain]

	for i, name := range req.ProductNames {
		price := 0.0
		available := false

		if shopExists {
			if p, found := shopData.Prices[name]; found {
				price = p
				available = true
			}
		}

		responses[i] = PriceResponse{
			ProductName: name,
			ShopChain:   req.ShopChain,
			Price:       price,
			Available:   available,
		}
	}

	return responses, nil
}

func (ps *PriceScraper) GetShopCoordinates(shopChain string) (Coordinates, error) {
	shopData, exists := ps.shops[shopChain]
	if !exists {
		return Coordinates{}, fmt.Errorf("shop chain %s not found in database", shopChain)
	}

	return shopData.Coordinates, nil
}

func (ps *PriceScraper) GetShopCoordinatesBulk(shopChains []string) (map[string]Coordinates, error) {
	coords := make(map[string]Coordinates, len(shopChains))
	for _, shopChain := range shopChains {
		shopData, exists := ps.shops[shopChain]
		if !exists {
			return nil, fmt.Errorf("shop chain %s not found in database", shopChain)
		}
		coords[shopChain] = shopData.Coordinates
	}
	return coords, nil
}

func (ps *PriceScraper) GetShopChains() []string {
	return ps.cachedChains
}

// SetShopsForTesting allows injecting shop data for testing purposes.
func (ps *PriceScraper) SetShopsForTesting(shops map[string]ShopData) {
	ps.shops = shops
	ps.updateCachedChains()
}

func (ps *PriceScraper) updateCachedChains() {
	ps.cachedChains = make([]string, 0, len(ps.shops))
	for chain := range ps.shops {
		ps.cachedChains = append(ps.cachedChains, chain)
	}
}
