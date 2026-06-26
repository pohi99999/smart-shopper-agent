package agents

import (
	"os"
	"testing"
)

func TestParser_Parse_MissingKey(t *testing.T) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	_, err := parser.Parse("veszek valamit")
	if err == nil {
		t.Fatalf("Expected an error due to missing API key, got nil")
	}
	if err.Error() != "GEMINI_API_KEY is not set or invalid" {
		t.Errorf("Expected GEMINI_API_KEY is not set or invalid error, got %v", err)
	}
}

func TestParser_Parse_Live_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-bound test in short mode")
	}
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "invalid_fake_key_123")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	_, err := parser.Parse("veszek valamit")
	if err == nil {
		t.Fatalf("Expected an error due to invalid API key, got nil")
	}
}
