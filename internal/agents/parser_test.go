package agents

import (
	"os"
	"testing"
)

func TestParser_Parse_Mock(t *testing.T) {
	// Let's force the parser to use the fallback mock by unsetting the API key
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	// It should return the hardcoded mock: tojás 10, kenyér 1
	list, err := parser.Parse("veszek valamit")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(list.Items) != 2 {
		t.Fatalf("Expected 2 items in fallback mock, got %d", len(list.Items))
	}

	if list.Items[0].Name != "tojás" || list.Items[0].Quantity != 10 {
		t.Errorf("Expected tojás (10), got %v", list.Items[0])
	}
	if list.Items[1].Name != "kenyér" || list.Items[1].Quantity != 1 {
		t.Errorf("Expected kenyér (1), got %v", list.Items[1])
	}
}

func TestParser_Parse_Live_Error(t *testing.T) {
	// Let's test with a fake API key to ensure it handles the error gracefully
	// and doesn't panic.
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "invalid_fake_key_123")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	// It should attempt an HTTP call and return an error (400 Bad Request usually for bad key)
	_, err := parser.Parse("veszek valamit")
	if err == nil {
		t.Fatalf("Expected an error due to invalid API key, got nil")
	}
}
