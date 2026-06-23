package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/api"
	"smart-shopper-agent/internal/mcp"
)

func main() {
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

	// 4. Register route
	http.HandleFunc("/api/v1/optimize", apiHandler.OptimizeHandler)
	http.HandleFunc("/api/v1/admin/prices", apiHandler.AdminPricesHandler)

	// 5. Start Server
	port := ":8080"
	slog.Info("Server is running", "port", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
