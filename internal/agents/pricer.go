package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

const fallbackPrice = 299.0

type Pricer struct {
	scraper *mcp.PriceScraper
}

func NewPricer(scraper *mcp.PriceScraper) *Pricer {
	return &Pricer{
		scraper: scraper,
	}
}

func (pr *Pricer) GetPrices(list models.ShoppingList) (map[string]float64, error) {
	chains := pr.scraper.GetShopChains()
	totals := make(map[string]float64)

	productNames := make([]string, len(list.Items))
	for i, item := range list.Items {
		productNames[i] = item.Name
	}

	for _, chain := range chains {
		respBatch, err := pr.scraper.ScrapePrices(mcp.PriceBatchRequest{
			ShopChain:    chain,
			ProductNames: productNames,
		})
		if err != nil {
			return nil, err
		}

		var total float64
		for i, resp := range respBatch {
			price := resp.Price
			if !resp.Available {
				price = fallbackPrice
			}
			total += price * float64(list.Items[i].Quantity)
		}
		totals[chain] = total
	}

	return totals, nil
}
