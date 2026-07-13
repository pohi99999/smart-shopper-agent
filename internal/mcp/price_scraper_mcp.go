package mcp

import (
	"encoding/json"
	"fmt"
	"os"
)

type PriceRequest struct {
	ProductName string `json:"product_name"`
	ShopChain   string `json:"shop_chain"`
}

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
	shops map[string]ShopData
}

func NewPriceScraper() *PriceScraper {
	ps := &PriceScraper{
		shops: make(map[string]ShopData),
	}

	// Olvassuk be a JSON fájlt egyszer, inicializáláskor
	data, err := os.ReadFile("internal/data/prices.json")
	if err == nil {
		// Ha nincs hiba, próbáljuk meg parse-olni
		_ = json.Unmarshal(data, &ps.shops)
	}

	return ps
}

func (ps *PriceScraper) ScrapePrice(req PriceRequest) (PriceResponse, error) {
	price := 299.0
	available := false

	if shopData, exists := ps.shops[req.ShopChain]; exists {
		if p, found := shopData.Prices[req.ProductName]; found {
			price = p
			available = true
		}
	}

	// Ha nem találtuk meg a terméket vagy a boltot, alapértelmezett árral térünk vissza a leírás alapján
	if !available {
		price = 299.0
		available = true
	}

	return PriceResponse{
		ProductName: req.ProductName,
		ShopChain:   req.ShopChain,
		Price:       price,
		Available:   available,
	}, nil
}

func (ps *PriceScraper) ScrapePrices(req PriceBatchRequest) ([]PriceResponse, error) {
	responses := make([]PriceResponse, len(req.ProductNames))
	shopData, shopExists := ps.shops[req.ShopChain]

	for i, name := range req.ProductNames {
		price := 299.0
		available := false

		if shopExists {
			if p, found := shopData.Prices[name]; found {
				price = p
				available = true
			}
		}

		if !available {
			price = 299.0
			available = true
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
