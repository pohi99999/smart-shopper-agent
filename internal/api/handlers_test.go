package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"smart-shopper-agent/internal/utils"
	"sync"
	"testing"
	"time"

	"smart-shopper-agent/internal/agents"
	"smart-shopper-agent/internal/mcp"
)

func TestAdminPricesHandler(t *testing.T) {
	t.Run("GET Server Configuration Error", func(t *testing.T) {
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "some-token")
		rec := httptest.NewRecorder()

		handler.AdminPricesGetHandler(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500 Internal Server Error, got %d", rec.Code)
		}
	})

	t.Run("Missing Token", func(t *testing.T) {
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "secret-admin-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		rec := httptest.NewRecorder()

		handler.AdminPricesGetHandler(rec, req)

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
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "secret-admin-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "invalid-token")
		rec := httptest.NewRecorder()

		handler.AdminPricesGetHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("Valid Token", func(t *testing.T) {
		utils.ResetPricesFilePathCacheForTesting()
		tempDir := t.TempDir()
		filePath := tempDir + "/prices.json"
		if err := os.WriteFile(filePath, []byte(`{"TestShop":{"coordinates":{"latitude":47.1234,"longitude":17.5678},"prices":{"tej":250}}}`), 0644); err != nil {
			t.Fatalf("Failed to create temp prices.json: %v", err)
		}

		utils.ResetPricesFilePathCacheForTesting()
		defer utils.ResetPricesFilePathCacheForTesting()
		originalEnv := os.Getenv("PRICES_FILE_PATH")
		os.Setenv("PRICES_FILE_PATH", filePath)
		defer os.Setenv("PRICES_FILE_PATH", originalEnv)
		utils.ResetPricesFilePathCacheForTesting()

		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "secret-admin-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/prices", nil)
		req.Header.Set("X-Admin-Token", "secret-admin-token-123")
		rec := httptest.NewRecorder()

		handler.AdminPricesGetHandler(rec, req)

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

	t.Run("POST Valid Token and Body", func(t *testing.T) {
		utils.ResetPricesFilePathCacheForTesting()
		tempDir := t.TempDir()
		filePath := tempDir + "/prices.json"
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create temp prices.json: %v", err)
		}

		utils.ResetPricesFilePathCacheForTesting()
		defer utils.ResetPricesFilePathCacheForTesting()
		originalEnv := os.Getenv("PRICES_FILE_PATH")
		os.Setenv("PRICES_FILE_PATH", filePath)
		defer os.Setenv("PRICES_FILE_PATH", originalEnv)
		utils.ResetPricesFilePathCacheForTesting()

		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "test-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		newPrices := map[string]interface{}{
			"TestShop": map[string]interface{}{
				"coordinates": map[string]float64{
					"latitude":  47.1234,
					"longitude": 17.5678,
				},
				"prices": map[string]float64{
					"tej": 250,
				},
			},
		}
		newJSON, _ := json.Marshal(newPrices)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prices", bytes.NewBuffer(newJSON))
		req.Header.Set("X-Admin-Token", "test-token-123")
		rec := httptest.NewRecorder()

		handler.AdminPricesPostHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		writtenData, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read updated prices.json: %v", err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(writtenData, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal written prices.json: %v", err)
		}
		if _, exists := decoded["TestShop"]; !exists {
			t.Errorf("Expected 'TestShop' to exist in written prices.json")
		}
	})

	t.Run("POST Invalid JSON", func(t *testing.T) {
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "test-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prices", bytes.NewBuffer([]byte(`{ invalid json }`)))
		req.Header.Set("X-Admin-Token", "test-token-123")
		rec := httptest.NewRecorder()

		handler.AdminPricesPostHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("POST Unauthorized", func(t *testing.T) {
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "test-token-123")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prices", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("X-Admin-Token", "wrong-token")
		rec := httptest.NewRecorder()

		handler.AdminPricesPostHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("POST Server Configuration Error", func(t *testing.T) {
		originalToken := os.Getenv("ADMIN_TOKEN")
		os.Setenv("ADMIN_TOKEN", "")
		defer os.Setenv("ADMIN_TOKEN", originalToken)

		handler := NewAPIHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prices", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("X-Admin-Token", "some-token")
		rec := httptest.NewRecorder()

		handler.AdminPricesPostHandler(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500 Internal Server Error, got %d", rec.Code)
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

	t.Run("Input Too Long", func(t *testing.T) {
		longInput := string(make([]byte, 2001))
		body, _ := json.Marshal(OptimizeRequest{UserInput: longInput})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("Failed to decode JSON error response: %v", err)
		}

		if errResp.Error != "Input too long" {
			t.Errorf("Expected error message 'Input too long', got '%s'", errResp.Error)
		}
	})

	t.Run("Request Body Too Large", func(t *testing.T) {
		largeBody := make([]byte, 1048577) // 1MB + 1 byte
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(largeBody))
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})
	t.Run("Invalid Latitude", func(t *testing.T) {
		reqBody := OptimizeRequest{
			UserInput: "tej",
			UserCoords: mcp.Coordinates{
				Latitude:  100.0,
				Longitude: 16.0,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("Failed to decode JSON error response: %v", err)
		}

		if errResp.Error != "Invalid latitude" {
			t.Errorf("Expected error message 'Invalid latitude', got '%s'", errResp.Error)
		}
	})

	t.Run("Invalid Longitude", func(t *testing.T) {
		reqBody := OptimizeRequest{
			UserInput: "tej",
			UserCoords: mcp.Coordinates{
				Latitude:  46.0,
				Longitude: 200.0,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimize", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		handler.OptimizeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("Failed to decode JSON error response: %v", err)
		}

		if errResp.Error != "Invalid longitude" {
			t.Errorf("Expected error message 'Invalid longitude', got '%s'", errResp.Error)
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

func TestSendJSONError(t *testing.T) {
	rec := httptest.NewRecorder()
	expectedMessage := "Test Error Message"
	expectedStatusCode := http.StatusBadRequest

	SendJSONError(rec, expectedMessage, expectedStatusCode)

	if rec.Code != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode JSON error response: %v", err)
	}

	if errResp.Error != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, errResp.Error)
	}

	if errResp.Code != expectedStatusCode {
		t.Errorf("Expected error code %d in JSON body, got %d", expectedStatusCode, errResp.Code)
	}
}

func TestOptimizeHandler_MethodNotAllowed(t *testing.T) {
	handler := NewAPIHandler(nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/optimize", nil)
	rec := httptest.NewRecorder()

	handler.OptimizeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestAPIHandler_getPricesData(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a valid JSON file
	validFile := tempDir + "/prices.json"
	validData := []byte(`{"test_shop": {"coordinates": {"lat": 47.0, "lon": 19.0}, "prices": {"milk": 2.0}}}`)
	if err := os.WriteFile(validFile, validData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create an invalid JSON file
	invalidFile := tempDir + "/invalid.json"
	if err := os.WriteFile(invalidFile, []byte(`{invalid_json`), 0644); err != nil {
		t.Fatalf("Failed to write invalid test file: %v", err)
	}

	tests := []struct {
		name      string
		setup     func()
		cleanup   func()
		wantError bool
		wantData  bool
	}{
		{
			name: "Success - Cache Miss (Reads from File)",
			setup: func() {
				os.Setenv("PRICES_FILE_PATH", validFile)
			},
			cleanup: func() {
				os.Unsetenv("PRICES_FILE_PATH")
				utils.ResetPricesFilePathCacheForTesting()
			},
			wantError: false,
			wantData:  true,
		},
		{
			name: "Error - File Not Found",
			setup: func() {
				os.Setenv("PRICES_FILE_PATH", tempDir+"/nonexistent.json")
			},
			cleanup: func() {
				os.Unsetenv("PRICES_FILE_PATH")
				utils.ResetPricesFilePathCacheForTesting()
			},
			wantError: true,
			wantData:  false,
		},
		{
			name: "Error - Invalid JSON",
			setup: func() {
				os.Setenv("PRICES_FILE_PATH", invalidFile)
			},
			cleanup: func() {
				os.Unsetenv("PRICES_FILE_PATH")
				utils.ResetPricesFilePathCacheForTesting()
			},
			wantError: true,
			wantData:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ResetPricesFilePathCacheForTesting()
			if tt.setup != nil {
				tt.setup()
			}
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup()
				}
			}()

			handler := NewAPIHandler(nil, nil, nil)
			data, err := handler.getPricesData()

			if tt.wantError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.wantData && data == nil {
				t.Errorf("Expected data but got nil")
			}
			if !tt.wantData && data != nil {
				t.Errorf("Expected no data but got %v", data)
			}
		})
	}

	t.Run("Success - Cache Hit", func(t *testing.T) {
		utils.ResetPricesFilePathCacheForTesting()
		os.Setenv("PRICES_FILE_PATH", validFile)
		defer func() {
			os.Unsetenv("PRICES_FILE_PATH")
			utils.ResetPricesFilePathCacheForTesting()
		}()

		handler := NewAPIHandler(nil, nil, nil)

		// First call - cache miss
		data1, err1 := handler.getPricesData()
		if err1 != nil {
			t.Fatalf("First call failed: %v", err1)
		}
		if data1 == nil {
			t.Fatalf("First call returned nil data")
		}

		// Intentionally break the file path so next call MUST use cache
		os.Setenv("PRICES_FILE_PATH", tempDir+"/nonexistent2.json")
		utils.ResetPricesFilePathCacheForTesting()

		// Second call - should hit cache and succeed
		data2, err2 := handler.getPricesData()
		if err2 != nil {
			t.Errorf("Second call (cache hit) failed: %v", err2)
		}

		// The data from cache should be the exact same object reference or value
		// Testing generic equality for unmarshaled json map
		if data2 == nil {
			t.Errorf("Second call returned nil data")
		}
	})
}

type failingResponseWriter struct {
	http.ResponseWriter
}

func (w *failingResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write error")
}

func (w *failingResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *failingResponseWriter) WriteHeader(statusCode int) {
}

func TestSendJSONError_ErrorPath(t *testing.T) {
	// Create a response writer that fails on write
	w := &failingResponseWriter{}

	// Call the function
	SendJSONError(w, "Test Error", http.StatusInternalServerError)

	// This test asserts nothing panic or crash when writing to connection fails
}

func TestAPIHandler_getPricesData_DoubleCheck(t *testing.T) {
	tempDir := t.TempDir()

	validFile := tempDir + "/prices.json"
	validData := []byte(`{"test_shop": {"coordinates": {"lat": 47.0, "lon": 19.0}, "prices": {"milk": 2.0}}}`)
	if err := os.WriteFile(validFile, validData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	utils.ResetPricesFilePathCacheForTesting()
	os.Setenv("PRICES_FILE_PATH", validFile)
	defer func() {
		os.Unsetenv("PRICES_FILE_PATH")
		utils.ResetPricesFilePathCacheForTesting()
	}()

	handler := NewAPIHandler(nil, nil, nil)

	var wg sync.WaitGroup

	// Block all goroutines at the first RLock check by holding the WLock
	handler.pricesCacheMut.Lock()

	numGoroutines := 100
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.getPricesData()
		}()
	}

	// Give them time to start and block on RLock
	time.Sleep(50 * time.Millisecond)

	// Release WLock. They all acquire RLock, see cache is nil, release RLock, and request WLock.
	// One wins, does file IO, sets cache.
	// The rest get WLock later and hit the double check.
	handler.pricesCacheMut.Unlock()

	wg.Wait()
}
