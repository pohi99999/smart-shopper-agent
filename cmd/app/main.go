package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/api"
	"smart-shopper-agent/internal/mcp"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "smart-shopper-agent/docs"
)

// @title Smart Shopper Agent API
// @version 1.0
// @description Backend API for the Smart Shopper Agent project.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load environment variables if .env file exists
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")
	// Configure JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Smart Shopper Agent API server is starting...")

	// 1. Initialize MCP tools
	scraper := mcp.NewPriceScraper()
	planner := mcp.NewRoutePlanner()
	slog.Info("Initialized MCP tools", "scraper", fmt.Sprintf("%T", scraper), "planner", fmt.Sprintf("%T", planner))

	// 2. Initialize Agents with injected dependencies
	parser := agents.NewParser()
	pricer := agents.NewPricer(scraper)
	optimizer := agents.NewOptimizer(planner, scraper)
	slog.Info("Initialized agents", "parser", fmt.Sprintf("%T", parser), "pricer", fmt.Sprintf("%T", pricer), "optimizer", fmt.Sprintf("%T", optimizer))

	// 3. Initialize API Handler
	apiHandler := api.NewAPIHandler(parser, pricer, optimizer)

	// Combine Middlewares
	// Both endpoints need security headers, but optimize also needs rate limiting

	optimizeHandler := api.SecurityHeadersMiddleware(api.RateLimitMiddleware(apiHandler.OptimizeHandler))
	adminPricesHandler := api.SecurityHeadersMiddleware(apiHandler.AdminPricesHandler)

	// 4. Register route
	http.HandleFunc("/api/v1/optimize", optimizeHandler)
	http.HandleFunc("/api/v1/admin/prices", adminPricesHandler)

	// Register Swagger route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// 5. Start Server
	port := ":8080"
	slog.Info("Server is running", "port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
