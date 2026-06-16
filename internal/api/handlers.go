package api

import (
	"encoding/json"
	"net/http"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
)

type APIHandler struct {
	parser    *agents.Parser
	pricer    *agents.Pricer
	optimizer *agents.Optimizer
}

func NewAPIHandler(parser *agents.Parser, pricer *agents.Pricer, optimizer *agents.Optimizer) *APIHandler {
	return &APIHandler{
		parser:    parser,
		pricer:    pricer,
		optimizer: optimizer,
	}
}

type OptimizeRequest struct {
	UserInput  string          `json:"user_input"`
	UserCoords mcp.Coordinates `json:"coords"`
}

type OptimizeResponse struct {
	RoutePlan models.RoutePlan `json:"route_plan"`
	TotalCost float64          `json:"total_cost"`
}

func (h *APIHandler) OptimizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Parser
	shoppingList, err := h.parser.Parse(req.UserInput)
	if err != nil {
		http.Error(w, "Parser error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Pricer
	prices, err := h.pricer.GetPrices(shoppingList)
	if err != nil {
		http.Error(w, "Pricer error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Optimizer
	routePlan, err := h.optimizer.Optimize(shoppingList, prices, req.UserCoords)
	if err != nil {
		http.Error(w, "Optimizer error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total cost
	var totalCost float64
	if len(routePlan.Steps) > 0 {
		bestShop := routePlan.Steps[0].ShopName
		totalCost = prices[bestShop]
	}

	resp := OptimizeResponse{
		RoutePlan: routePlan,
		TotalCost: totalCost,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
