package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

type Pricer struct {
	scraper *mcp.PriceScraper
}

func NewPricer(scraper *mcp.PriceScraper) *Pricer {
	return &Pricer{
		scraper: scraper,
	}
}

func (pr *Pricer) GetPrices(list models.ShoppingList) (map[string]float64, error) {
	chains := []string{"Aldi", "Interspar"}
	totals := make(map[string]float64)

	for _, chain := range chains {
		var total float64
		for _, item := range list.Items {
			resp, err := pr.scraper.ScrapePrice(mcp.PriceRequest{
				ProductName: item.Name,
				ShopChain:   chain,
			})
			if err != nil {
				return nil, err
			}
			if resp.Available {
				total += resp.Price * float64(item.Quantity)
			}
		}
		totals[chain] = total
	}

	return totals, nil
}
