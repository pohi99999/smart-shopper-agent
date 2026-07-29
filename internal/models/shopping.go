package models

type ShoppingItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type ShoppingList struct {
	Items []ShoppingItem `json:"items"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type RouteStep struct {
	ShopName    string      `json:"shop_name"`
	Items       []string    `json:"items"`
	Coordinates Coordinates `json:"coordinates"`
}

type RoutePlan struct {
	Steps []RouteStep `json:"steps"`
}
