package agents

import (
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
