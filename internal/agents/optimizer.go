package agents

import (
	"fmt"

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
	if len(prices) == 0 {
		return models.RoutePlan{}, fmt.Errorf("no shops found within 50 km")
	}

	bestShop := ""
	minCost := -1.0

	// Gather all destinations
	destinations := make(map[string]mcp.Coordinates)
	for shopName := range prices {
		coords, err := o.scraper.GetShopCoordinates(shopName)
		if err != nil {
			return models.RoutePlan{}, err
		}
		destinations[shopName] = coords
	}

	matrixReq := mcp.RouteMatrixRequest{
		Source:       userCoords,
		Destinations: destinations,
	}

	matrixResp, err := o.planner.CalculateRouteMatrix(matrixReq)
	if err != nil {
		return models.RoutePlan{}, err
	}

	for shopName, price := range prices {
		routeResp, ok := matrixResp[shopName]
		if !ok {
			// This shouldn't happen unless OSRM failed to return a route for this specific shop,
			// but we'll safely skip if it happens.
			continue
		}

		// Skip if the distance is greater than 50 km
		if routeResp.DistanceKM > 50.0 {
			continue
		}

		if minCost == -1.0 || price < minCost {
			minCost = price
			bestShop = shopName
		}
	}

	if bestShop == "" {
		return models.RoutePlan{}, fmt.Errorf("no shops found within 50 km")
	}

	items := make([]string, 0, len(list.Items))
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
