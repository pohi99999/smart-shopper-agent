package agents

import (
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

const OptimizerSystemPrompt = "You are an expert route and budget optimizer that selects the best shopping destination based on prices and travel distance."

type Optimizer struct {
	planner *mcp.RoutePlanner
}

func NewOptimizer(planner *mcp.RoutePlanner) *Optimizer {
	return &Optimizer{
		planner: planner,
	}
}

func (o *Optimizer) Optimize(list models.ShoppingList, prices map[string]float64, userCoords mcp.Coordinates) (models.RoutePlan, error) {
	// Boltok koordinátái
	shopCoords := map[string]mcp.Coordinates{
		"Aldi":      {Latitude: 47.4800, Longitude: 19.0600},
		"Interspar": {Latitude: 47.5100, Longitude: 19.0200},
	}

	bestShop := ""
	minCost := -1.0

	for shopName, price := range prices {
		coords := shopCoords[shopName]
		_, err := o.planner.CalculateRoute(mcp.RouteRequest{
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

