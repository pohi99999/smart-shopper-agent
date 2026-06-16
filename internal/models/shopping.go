package models

type ShoppingItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type ShoppingList struct {
	Items []ShoppingItem `json:"items"`
}

type RouteStep struct {
	ShopName string   `json:"shop_name"`
	Items    []string `json:"items"`
}

type RoutePlan struct {
	Steps []RouteStep `json:"steps"`
}
