package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/mcp"
	"testing"
)

func TestAdminPricesHandler(t *testing.T) {
	handler := NewAPIHandler(nil, nil, nil)

	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		rec := httptest.NewRecorder()

		handler.AdminPricesHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("Failed to decode JSON error response: %v", err)
		}

		if errResp.Error != "Unauthorized" {
			t.Errorf("Expected 'Unauthorized' error message, got %s", errResp.Error)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "invalid-token")
		rec := httptest.NewRecorder()

		handler.AdminPricesHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Valid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "secret-admin-token-123")
		rec := httptest.NewRecorder()

		handler.AdminPricesHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if resp["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", resp["status"])
		}
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "secret-admin-token-123")
		rec := httptest.NewRecorder()

		handler.AdminPricesHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})
}

func TestOptimizeHandler_InvalidMethodAndBody(t *testing.T) {
	handler := NewAPIHandler(nil, nil, nil)

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/optimize", nil)
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405 Method Not Allowed, got %d", rec.Code)
		}
	})

	t.Run("Invalid Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer([]byte("invalid json")))
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})
}

func TestOptimizeHandler_Integration(t *testing.T) {
	// Creating full instances to test if they string together without panic
	// Note: We're not doing fully mocked endpoints here to keep it simple,
	// just verifying the structure works. (The parser might use the fallback, which is fine)

	scraper := mcp.NewPriceScraper()
	planner := mcp.NewRoutePlanner()
	parser := agents.NewParser()
	pricer := agents.NewPricer(scraper)
	optimizer := agents.NewOptimizer(planner, scraper)

	handler := NewAPIHandler(parser, pricer, optimizer)

	reqBody := OptimizeRequest{
		UserInput: "kenyér és tej",
		UserCoords: mcp.Coordinates{
			Latitude:  46.8400, // Zalaegerszeg
			Longitude: 16.8439,
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(jsonData))
	rec := httptest.NewRecorder()

	handler.OptimizeHandler(rec, req)

	// In test env, RoutePlanner might fail due to no OSRM mock if internet is down, or succeed.
	// So we don't strictly test for 200, but we test for NOT a panic and valid JSON response

	var errResp ErrorResponse
	var successResp OptimizeResponse

	// Either it's a 500 error struct OR 200 success struct
	if rec.Code == http.StatusOK {
		if err := json.NewDecoder(rec.Body).Decode(&successResp); err != nil {
			t.Fatalf("Failed to decode success JSON: %v", err)
		}
	} else {
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("Failed to decode error JSON: %v, Body: %s", err, rec.Body.String())
		}
	}
}

func TestOptimizeHandler_ParserError(t *testing.T) {
	// Let's force a parser error by setting invalid API key
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "invalid_key")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	scraper := mcp.NewPriceScraper()
	planner := mcp.NewRoutePlanner()
	parser := agents.NewParser()
	pricer := agents.NewPricer(scraper)
	optimizer := agents.NewOptimizer(planner, scraper)

	handler := NewAPIHandler(parser, pricer, optimizer)

	reqBody := OptimizeRequest{
		UserInput: "kenyér és tej",
		UserCoords: mcp.Coordinates{
			Latitude:  46.8400,
			Longitude: 16.8439,
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(jsonData))
	rec := httptest.NewRecorder()

	handler.OptimizeHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", rec.Code)
	}
}
