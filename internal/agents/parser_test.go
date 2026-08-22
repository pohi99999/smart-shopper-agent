package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
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
	parser.APIKey = ""

	_, err := parser.Parse(context.Background(), "veszek valamit")
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
	parser.APIKey = "invalid_fake_key_123"

	_, err := parser.Parse(context.Background(), "veszek valamit")
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
	parser.APIKey = "test_mock_api_key"
	parser.Client = mockClient

	result, err := parser.Parse(context.Background(), "buy 1 milk")
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
	parser.APIKey = "dummy_key"

	_, err := parser.Parse(context.Background(), "veszek valamit")
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

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatalf("Expected NewParser to return a non-nil pointer")
	}
	if parser.Client == nil {
		t.Fatalf("Expected parser.Client to be initialized")
	}
	if parser.Client.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", parser.Client.Timeout)
	}
	expectedURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	if parser.APIURL != expectedURL {
		t.Errorf("Expected APIURL %q, got %q", expectedURL, parser.APIURL)
	}
}

func TestParser_Parse_NetworkError(t *testing.T) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "dummy_key")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	originalTransport := http.DefaultTransport
	http.DefaultTransport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("simulated network error")
		},
	}
	defer func() { http.DefaultTransport = originalTransport }()

	parser := NewParser()
	parser.APIKey = "dummy_key"

	// Fast fail for retries
	parser.Client.Timeout = 1 * time.Millisecond

	_, err := parser.Parse(context.Background(), "buy 1 milk")
	if err == nil {
		t.Fatalf("Expected an error due to network failure, got nil")
	}
	if !strings.Contains(err.Error(), "Gemini API network error") {
		t.Errorf("Expected 'Gemini API network error', got %v", err)
	}
}

func TestParser_doAttempt_BadURL(t *testing.T) {
	parser := NewParser()
	badURL := string([]byte{0x7f}) // Invalid control character for URL
	_, err := parser.doAttempt(context.Background(), http.DefaultClient, badURL, "dummy_key", []byte(`{}`), 0)
	if err == nil {
		t.Fatalf("Expected an error due to invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create HTTP request") {
		t.Errorf("Expected 'failed to create HTTP request' error, got %v", err)
	}
}

func TestParser_doAttempt_BadStatus(t *testing.T) {
	mockClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			Header:     make(http.Header),
		}
	})

	parser := NewParser()
	_, err := parser.doAttempt(context.Background(), mockClient, "http://dummy", "dummy_key", []byte(`{}`), 0)
	if err == nil {
		t.Fatalf("Expected an error due to bad status code, got nil")
	}
	if !strings.Contains(err.Error(), "API request failed with status code 500") {
		t.Errorf("Expected 'API request failed with status code 500' error, got %v", err)
	}
}

func TestParser_doAttempt_EmptyCandidates(t *testing.T) {
	mockClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"candidates":[]}`)),
			Header:     make(http.Header),
		}
	})

	parser := NewParser()
	_, err := parser.doAttempt(context.Background(), mockClient, "http://dummy", "dummy_key", []byte(`{}`), 0)
	if err == nil {
		t.Fatalf("Expected an error due to empty candidates, got nil")
	}
	if !strings.Contains(err.Error(), "invalid or empty response from Gemini API") {
		t.Errorf("Expected 'invalid or empty response from Gemini API' error, got %v", err)
	}
}

func TestParser_doAttempt_BadShoppingListJSON(t *testing.T) {
	mockClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"candidates":[{"content":{"parts":[{"text":"invalid_json"}]}}]}`)),
			Header:     make(http.Header),
		}
	})

	parser := NewParser()
	_, err := parser.doAttempt(context.Background(), mockClient, "http://dummy", "dummy_key", []byte(`{}`), 0)
	if err == nil {
		t.Fatalf("Expected an error due to bad shopping list JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse shopping list JSON") {
		t.Errorf("Expected 'failed to parse shopping list JSON' error, got %v", err)
	}
}

func TestParser_Parse_DefaultClientAndURL(t *testing.T) {
	parser := NewParser()
	parser.Client = nil
	parser.APIURL = ""

	// We'll give it a fake key so it fails on auth if it gets that far.
	parser.APIKey = "fake_key_for_defaults"

	_, err := parser.Parse(context.Background(), "veszek valamit")
	if err == nil {
		t.Fatalf("Expected an error due to fake key or network, got nil")
	}
}
