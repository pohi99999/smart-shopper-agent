package agents

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"sync"

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
	var mu sync.Mutex
	var g errgroup.Group

	for shopName, price := range prices {
		shopName := shopName
		price := price

		g.Go(func() error {
			coords, err := o.scraper.GetShopCoordinates(shopName)
			if err != nil {
				return err
			}
			routeResp, err := o.planner.CalculateRoute(mcp.RouteRequest{
				Source:      userCoords,
				Destination: coords,
			})
			if err != nil {
				return err
			}

			// Skip if the distance is greater than 50 km
			if routeResp.DistanceKM > 50.0 {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()
			if minCost == -1.0 || price < minCost {
				minCost = price
				bestShop = shopName
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return models.RoutePlan{}, err
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
