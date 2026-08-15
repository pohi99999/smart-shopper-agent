package agents

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

func BenchmarkParser_Parse_FastFail(b *testing.B) {
	// Setup with invalid key to fail fast and measure just the setup phase
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "your_api_key_here")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse("buy 1 milk")
	}
}

func BenchmarkParser_Parse_Success(b *testing.B) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "valid_key")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()
	parser.APIKey = "valid_key"
	parser.Client = NewTestClient(func(req *http.Request) *http.Response {
		responseJSON := `{
			"candidates": [
				{
					"content": {
						"parts": [
							{
								"text": "{\"items\": [{\"name\": \"milk\", \"quantity\": 1, \"unit\": \"\"}]}"
							}
						]
					}
				}
			]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(responseJSON)),
			Header:     make(http.Header),
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse("buy 1 milk")
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
