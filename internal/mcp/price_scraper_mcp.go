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

type ShopData struct {
	Coordinates Coordinates        `json:"coordinates"`
	Prices      map[string]float64 `json:"prices"`
}

type PriceScraper struct{}

func NewPriceScraper() *PriceScraper {
	return &PriceScraper{}
}

func (ps *PriceScraper) ScrapePrice(req PriceRequest) (PriceResponse, error) {
	// Olvassuk be a JSON fájlt a megadott helyről
	data, err := os.ReadFile("internal/data/prices.json")
	if err != nil {
		// Ha hiba van a beolvasáskor (pl. nincs fájl), alapértelmezett árral térünk vissza
		return PriceResponse{
			ProductName: req.ProductName,
			ShopChain:   req.ShopChain,
			Price:       299.0,
			Available:   true,
		}, nil
	}

	var shops map[string]ShopData
	if err := json.Unmarshal(data, &shops); err != nil {
		// Ha nem sikerült parse-olni, szintén alapértelmezett ár
		return PriceResponse{
			ProductName: req.ProductName,
			ShopChain:   req.ShopChain,
			Price:       299.0,
			Available:   true,
		}, nil
	}

	price := 299.0
	available := false

	if shopData, exists := shops[req.ShopChain]; exists {
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

func (ps *PriceScraper) GetShopCoordinates(shopChain string) (Coordinates, error) {
	data, err := os.ReadFile("internal/data/prices.json")
	if err != nil {
		return Coordinates{}, err
	}

	var shops map[string]ShopData
	if err := json.Unmarshal(data, &shops); err != nil {
		return Coordinates{}, err
	}

	shopData, exists := shops[shopChain]
	if !exists {
		return Coordinates{}, fmt.Errorf("shop chain %s not found in database", shopChain)
	}

	return shopData.Coordinates, nil
}
