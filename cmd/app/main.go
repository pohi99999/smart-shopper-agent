package main

import (
	"fmt"
	"log"
	"net/http"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/api"
	"smart-shopper-agent/internal/mcp"
)

func main() {
	fmt.Println("Smart Shopper Agent API server is starting...")

	// 1. Initialize MCP tools
	scraper := mcp.NewPriceScraper()
	planner := mcp.NewRoutePlanner()
	fmt.Printf("Initialized MCP tools: PriceScraper (%T), RoutePlanner (%T)\n", scraper, planner)

	// 2. Initialize Agents with injected dependencies
	parser := agents.NewParser()
	pricer := agents.NewPricer(scraper)
	optimizer := agents.NewOptimizer(planner, scraper)
	fmt.Printf("Initialized agents: Parser (%T), Pricer (%T), Optimizer (%T)\n", parser, pricer, optimizer)

	// 3. Initialize API Handler
	apiHandler := api.NewAPIHandler(parser, pricer, optimizer)

	// 4. Register route
	http.HandleFunc("/api/v1/optimize", apiHandler.OptimizeHandler)
	http.HandleFunc("/api/v1/admin/prices", apiHandler.AdminPricesHandler)

	// 5. Start Server
	port := ":8080"
	fmt.Printf("Server is running on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
