package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

const OptimizerSystemPrompt = "You are an expert route and budget optimizer that selects the best shopping destination based on prices and travel distance."

type Optimizer struct {
	planner *mcp.RoutePlanner
	scraper *mcp.PriceScraper
}

func NewOptimizer(planner *mcp.RoutePlanner, scraper *mcp.PriceScraper) *Optimizer {
	return &Optimizer{
		planner: planner,
		scraper: scraper,
	}
}

func (o *Optimizer) Optimize(list models.ShoppingList, prices map[string]float64, userCoords mcp.Coordinates) (models.RoutePlan, error) {
	bestShop := ""
	minCost := -1.0

	for shopName, price := range prices {
		coords, err := o.scraper.GetShopCoordinates(shopName)
		if err != nil {
			return models.RoutePlan{}, err
		}
		_, err = o.planner.CalculateRoute(mcp.RouteRequest{
			Source:      userCoords,
			Destination: coords,
		})
		if err != nil {
			return models.RoutePlan{}, err
		}

		if minCost == -1.0 || price < minCost {
			minCost = price
			bestShop = shopName
		}
	}

	var items []string
	for _, item := range list.Items {
		items = append(items, item.Name)
	}

	step := models.RouteStep{
		ShopName: bestShop,
		Items:    items,
	}

	return models.RoutePlan{
		Steps: []models.RouteStep{step},
	}, nil
}

