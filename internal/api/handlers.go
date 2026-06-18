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

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func SendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  statusCode,
	})
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
		SendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Parser
	shoppingList, err := h.parser.Parse(req.UserInput)
	if err != nil {
		SendJSONError(w, "Parser error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Pricer
	prices, err := h.pricer.GetPrices(shoppingList)
	if err != nil {
		SendJSONError(w, "Pricer error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Optimizer
	routePlan, err := h.optimizer.Optimize(shoppingList, prices, req.UserCoords)
	if err != nil {
		SendJSONError(w, "Optimizer error: "+err.Error(), http.StatusInternalServerError)
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
		SendJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) AdminPricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("X-Admin-Token")
	if token != "secret-admin-token-123" {
		SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// This is a stub implementation. In a real application, you would
	// fetch the prices from a database or memory.
	prices := map[string]interface{}{
		"status": "success",
		"data": map[string]map[string]float64{
			"Aldi": {
				"tej":    300,
				"kenyer": 400,
			},
			"Interspar": {
				"tej":    350,
				"kenyer": 380,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prices); err != nil {
		SendJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
