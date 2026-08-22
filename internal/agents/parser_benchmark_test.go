package agents

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func BenchmarkParser_Parse_FastFail(b *testing.B) {
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "your_api_key_here")
	defer os.Setenv("GEMINI_API_KEY", originalAPIKey)

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(context.Background(), "buy 1 milk")
	}
}

func BenchmarkParser_Parse_ClientDisconnect(b *testing.B) {
	parser := NewParser()
	parser.APIKey = "fake_key"
	parser.APIURL = "http://dummy"
	parser.Client = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

		done := make(chan struct{})
		go func() {
			// In the optimized version we will pass ctx to Parse
			_, _ = parser.Parse(ctx, "buy 1 milk")
			close(done)
		}()

		select {
		case <-ctx.Done():
		case <-done:
		}

		<-done
		cancel()
	}
}
