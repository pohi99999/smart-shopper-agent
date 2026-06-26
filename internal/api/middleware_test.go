package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Create a dummy handler that returns HTTP 200 OK.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the dummy handler with our middleware.
	handler := SecurityHeadersMiddleware(dummyHandler)

	t.Run("Headers are set correctly on standard request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Assert status code
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Assert expected headers
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, DELETE",
			"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Token",
			"X-Content-Type-Options":       "nosniff",
		}

		for key, expectedVal := range expectedHeaders {
			actualVal := rec.Header().Get(key)
			if actualVal != expectedVal {
				t.Errorf("Expected header %s to be '%s', got '%s'", key, expectedVal, actualVal)
			}
		}
	})

	t.Run("OPTIONS request returns 200 OK immediately", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Assert status code is 200
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Assert expected headers are also present for OPTIONS
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, DELETE",
			"Access-Control-Allow-Headers": "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Token",
			"X-Content-Type-Options":       "nosniff",
		}

		for key, expectedVal := range expectedHeaders {
			actualVal := rec.Header().Get(key)
			if actualVal != expectedVal {
				t.Errorf("Expected header %s to be '%s', got '%s'", key, expectedVal, actualVal)
			}
		}
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	// Create a dummy handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Use a fresh rate limiter specifically for this test to avoid state issues
	// We'll temporarily replace the global limiter in middleware.go just for this test block if possible
	// Alternatively, we just hit it directly. The default is 10 requests / min.
	// Since we can't easily mock the global var here, we'll just test the logic

	handler := RateLimitMiddleware(dummyHandler)

	t.Run("Allows requests within limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234" // unique IP for this test

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", rec.Code)
		}
	})

	t.Run("Handles X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.1") // This IP will be used

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", rec.Code)
		}
	})

	t.Run("Blocks requests over limit", func(t *testing.T) {
		// Use a specific IP that will hit the rate limit
		ip := "10.0.0.1"

		// Send 10 requests which should be allowed (burst limit is 10)
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip + ":1234"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("Request %d failed unexpectedly: got %d", i+1, rec.Code)
			}
		}

		// The 11th request should be blocked
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip + ":1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429 Too Many Requests, got %d", rec.Code)
		}
	})
}
