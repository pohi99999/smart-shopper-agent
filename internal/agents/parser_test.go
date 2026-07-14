package agents

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
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

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestParser_Parse_BadJSONResponse(t *testing.T) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "dummy_key")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	originalTransport := http.DefaultTransport
	http.DefaultTransport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"invalid: json`)),
				Header:     make(http.Header),
			}, nil
		},
	}
	defer func() { http.DefaultTransport = originalTransport }()

	parser := NewParser()

	_, err := parser.Parse("veszek valamit")
	if err == nil {
		t.Fatalf("Expected an error due to bad JSON response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected 'failed to decode response' error, got %v", err)
	}
}
