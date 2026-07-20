package agents

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{Transport: RoundTripFunc(fn)}
}

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

func TestParser_Parse_Success(t *testing.T) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "test_mock_api_key")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	mockResponseJSON := `{"candidates":[{"content":{"parts":[{"text":"{\"items\": [{\"name\": \"milk\", \"quantity\": 1}]}"}]}}]}`

	mockClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponseJSON)),
			Header:     make(http.Header),
		}
	})

	parser := NewParser()
	parser.Client = mockClient

	result, err := parser.Parse("buy 1 milk")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Name != "milk" {
		t.Errorf("Expected item name 'milk', got %s", result.Items[0].Name)
	}
	if result.Items[0].Quantity != 1 {
		t.Errorf("Expected item quantity 1, got %d", result.Items[0].Quantity)
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

func TestBuildRequestBody(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"normal text", "buy 2 apples"},
		{"empty string", ""},
		{"special characters", "milk & honey @ $5!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildRequestBody(tt.input)
			if err != nil {
				t.Fatalf("buildRequestBody() error = %v", err)
			}

			var reqBody GeminiRequest
			if err := json.Unmarshal(result, &reqBody); err != nil {
				t.Fatalf("failed to unmarshal JSON result: %v", err)
			}

			if len(reqBody.Contents) != 1 || len(reqBody.Contents[0].Parts) != 1 {
				t.Fatalf("expected 1 content part")
			}
			if reqBody.Contents[0].Parts[0].Text != tt.input {
				t.Errorf("expected text %q, got %q", tt.input, reqBody.Contents[0].Parts[0].Text)
			}

			if len(reqBody.SystemInstruction.Parts) != 1 {
				t.Fatalf("expected 1 system instruction part")
			}
			if reqBody.SystemInstruction.Parts[0].Text != ParserSystemPrompt {
				t.Errorf("expected system prompt %q, got %q", ParserSystemPrompt, reqBody.SystemInstruction.Parts[0].Text)
			}

			if reqBody.GenerationConfig.ResponseMimeType != "application/json" {
				t.Errorf("expected response mime type 'application/json', got %q", reqBody.GenerationConfig.ResponseMimeType)
			}
		})
	}
}
