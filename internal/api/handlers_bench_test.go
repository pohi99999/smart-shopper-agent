package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"smart-shopper-agent/internal/utils"
	"testing"
)

func BenchmarkAdminPricesGetHandler(b *testing.B) {
	// Setup
	os.Setenv("ADMIN_TOKEN", "test-token")
	defer os.Unsetenv("ADMIN_TOKEN")

	utils.ResetPricesFilePathCacheForTesting()
	defer utils.ResetPricesFilePathCacheForTesting()

	tmpDir, err := os.MkdirTemp("", "bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pricesPath := filepath.Join(tmpDir, "prices.json")
	os.Setenv("PRICES_FILE_PATH", pricesPath)
	defer os.Unsetenv("PRICES_FILE_PATH")

	// Let's create a reasonably large dummy data file to make the parsing overhead visible
	dummyData := make(map[string]map[string]float64)
	for i := 0; i < 100; i++ {
		shop := make(map[string]float64)
		for j := 0; j < 1000; j++ {
			// e.g. "item-42": 12.34
			shop["item"] = 12.34
		}
		dummyData["shop"] = shop
	}

	data, _ := json.Marshal(dummyData)
	if err := os.WriteFile(pricesPath, data, 0644); err != nil {
		b.Fatalf("failed to write prices: %v", err)
	}

	handler := NewAPIHandler(nil, nil, nil)

	req, _ := http.NewRequest(http.MethodGet, "/admin/prices", nil)
	req.Header.Set("X-Admin-Token", "test-token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.AdminPricesGetHandler(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}
	}
}
