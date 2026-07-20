package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/mcp"
	"smart-shopper-agent/internal/models"
	"sync"
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
	UserInput  string          `json:"user_input" example:"10 tojás és egy kenyér"`
	UserCoords mcp.Coordinates `json:"coords"`
}

type OptimizeResponse struct {
	RoutePlan models.RoutePlan `json:"route_plan"`
	TotalCost float64          `json:"total_cost" example:"1250"`
}

var (
	pricesFilePath     string
	pricesFilePathOnce sync.Once
)

func getPricesFilePath() string {
	pricesFilePathOnce.Do(func() {
		filePath := "internal/data/prices.json"
		if _, err := os.Stat(filePath); err != nil {
			if _, err2 := os.Stat("../../internal/data/prices.json"); err2 == nil {
				filePath = "../../internal/data/prices.json"
			}
		}
		pricesFilePath = filePath
	})
	return pricesFilePath
}

// OptimizeHandler godoc
// @Summary Calculate optimized shopping route
// @Description Extracts shopping items from natural language, fetches prices, and calculates the optimal shopping route and total cost based on the user's location.
// @Tags optimize
// @Accept json
// @Produce json
// @Param request body OptimizeRequest true "User input and coordinates"
// @Success 200 {object} OptimizeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /optimize [post]
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

// AdminPricesGetHandler godoc
// @Summary Fetch shop prices
// @Description Fetches shop prices. Requires an X-Admin-Token header.
// @Tags admin
// @Produce json
// @Param X-Admin-Token header string true "Admin Token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/prices [get]
func (h *APIHandler) AdminPricesGetHandler(w http.ResponseWriter, r *http.Request) {
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		SendJSONError(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("X-Admin-Token")
	expectedTokenHash := sha256.Sum256([]byte(adminToken))
	providedTokenHash := sha256.Sum256([]byte(token))

	if subtle.ConstantTimeCompare(providedTokenHash[:], expectedTokenHash[:]) != 1 {
		SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filePath := getPricesFilePath()
	bodyBytes, err := os.ReadFile(filePath)
	if err != nil {
		SendJSONError(w, "Failed to read prices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		SendJSONError(w, "Failed to parse prices data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	prices := map[string]interface{}{
		"status": "success",
		"data":   data,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prices); err != nil {
		SendJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AdminPricesPostHandler godoc
// @Summary Update shop prices
// @Description Updates shop prices. Requires an X-Admin-Token header.
// @Tags admin
// @Accept json
// @Produce json
// @Param X-Admin-Token header string true "Admin Token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/prices [post]
func (h *APIHandler) AdminPricesPostHandler(w http.ResponseWriter, r *http.Request) {
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		SendJSONError(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("X-Admin-Token")
	expectedTokenHash := sha256.Sum256([]byte(adminToken))
	providedTokenHash := sha256.Sum256([]byte(token))

	if subtle.ConstantTimeCompare(providedTokenHash[:], expectedTokenHash[:]) != 1 {
		SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var temp interface{}
	if err := json.Unmarshal(bodyBytes, &temp); err != nil {
		SendJSONError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	filePath := getPricesFilePath()
	if err := os.WriteFile(filePath, bodyBytes, 0644); err != nil {
		SendJSONError(w, "Failed to save prices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Prices updated successfully",
	})
}
