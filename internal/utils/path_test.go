package utils

import (
	"os"
	"testing"
)

func TestGetPricesFilePath(t *testing.T) {
	// Save the original working directory
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	t.Run("File exists in current directory (internal/data/prices.json)", func(t *testing.T) {
		ResetPricesFilePathCacheForTesting()
		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}
		defer os.Chdir(originalWD)

		dataDir := "internal/data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}

		filePath := dataDir + "/prices.json"
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		got := GetPricesFilePath()
		expected := "internal/data/prices.json"
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("File exists in parent directory (../../internal/data/prices.json)", func(t *testing.T) {
		ResetPricesFilePathCacheForTesting()
		tempDir := t.TempDir()

		// Create the target file structure
		dataDir := tempDir + "/internal/data"
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}

		filePath := dataDir + "/prices.json"
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create and switch to the API directory structure
		apiDir := tempDir + "/internal/api"
		if err := os.MkdirAll(apiDir, 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}

		if err := os.Chdir(apiDir); err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}
		defer os.Chdir(originalWD)

		got := GetPricesFilePath()
		expected := "../../internal/data/prices.json"
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})

	t.Run("File does not exist (fallback to default)", func(t *testing.T) {
		ResetPricesFilePathCacheForTesting()
		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}
		defer os.Chdir(originalWD)

		got := GetPricesFilePath()
		expected := "internal/data/prices.json"
		if got != expected {
			t.Errorf("Expected %q, got %q", expected, got)
		}
	})
}
