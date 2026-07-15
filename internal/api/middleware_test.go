package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Keep the original limiter to restore it later
	originalLimiter := limiter
	defer func() {
		limiter = originalLimiter
	}()

	// Reset the global limiter for testing
	limiter = NewRateLimiter(rate.Every(time.Minute/10), 2) // Allow max 2 burst

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := RateLimitMiddleware(nextHandler)

	t.Run("Normal requests are allowed", func(t *testing.T) {
		// First request: Should pass
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		rec1 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec1.Code)
		}

		// Second request: Should pass
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.1:1234"
		rec2 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec2.Code)
		}
	})

	t.Run("Exceeding rate limit returns 429", func(t *testing.T) {
		// Third request: Should be rate limited
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req3.RemoteAddr = "192.168.1.1:1234"
		rec3 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec3, req3)
		if rec3.Code != http.StatusTooManyRequests {
			t.Errorf("Expected 429 Too Many Requests, got %d", rec3.Code)
		}
	})

	t.Run("Different IP is not affected by another IP's rate limit", func(t *testing.T) {
		// Fourth request from a different IP: Should pass
		req4 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req4.RemoteAddr = "192.168.1.2:1234"
		rec4 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec4, req4)
		if rec4.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec4.Code)
		}
	})
}

func TestRateLimitMiddleware_ForwardedFor(t *testing.T) {
	// Keep the original limiter to restore it later
	originalLimiter := limiter
	defer func() {
		limiter = originalLimiter
	}()

	// Reset the global limiter for testing
	limiter = NewRateLimiter(rate.Every(time.Minute/10), 1) // Allow max 1 burst

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := RateLimitMiddleware(nextHandler)

	t.Run("Uses X-Forwarded-For instead of RemoteAddr", func(t *testing.T) {
		// First request using X-Forwarded-For: Should pass
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		req1.Header.Set("X-Forwarded-For", "10.0.0.1")
		rec1 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec1.Code)
		}

		// Second request using same X-Forwarded-For but different RemoteAddr: Should be rate limited
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.2:1234" // different remote addr, but same forwarded
		req2.Header.Set("X-Forwarded-For", "10.0.0.1")
		rec2 := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusTooManyRequests {
			t.Errorf("Expected 429 Too Many Requests, got %d", rec2.Code)
		}
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		expected   string
	}{
		{"No X-Forwarded-For", "192.168.1.100:1234", "", "192.168.1.100"},
		{"No port in RemoteAddr", "192.168.1.100", "", "192.168.1.100"},
		{"Spoofed + Real IP", "10.0.0.1:1234", "8.8.8.8, 203.0.113.5", "203.0.113.5"},
		{"All private IPs", "10.0.0.1:1234", "192.168.1.50, 10.0.0.2", "192.168.1.50"},
		{"Multiple public IPs", "10.0.0.1:1234", "203.0.113.1, 203.0.113.2", "203.0.113.2"},
		{"Invalid IPs in XFF", "10.0.0.1:1234", "invalid, 203.0.113.1, not-an-ip", "203.0.113.1"},
		{"Only invalid IPs", "10.0.0.1:1234", "invalid, not-an-ip", "10.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			actual := GetClientIP(req)
			if actual != tt.expected {
				t.Errorf("GetClientIP() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	// A dummy handler to wrap
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		allowedOrigin  string // Value to set in ALLOWED_ORIGIN env var
		expectedOrigin string // Expected value in Access-Control-Allow-Origin header
	}{
		{
			name:           "Default origin when env var is not set",
			allowedOrigin:  "",
			expectedOrigin: "",
		},
		{
			name:           "Custom origin when env var is set",
			allowedOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original env var and ensure it's restored after the test
			originalOrigin := os.Getenv("ALLOWED_ORIGIN")
			defer os.Setenv("ALLOWED_ORIGIN", originalOrigin)

			if tt.allowedOrigin == "" {
				os.Unsetenv("ALLOWED_ORIGIN")
			} else {
				os.Setenv("ALLOWED_ORIGIN", tt.allowedOrigin)
			}

			// Create the middleware wrapped handler
			handler := SecurityHeadersMiddleware(dummyHandler)

			// Create a request and a response recorder
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)

			// Check the Access-Control-Allow-Origin header
			actualOrigin := rr.Header().Get("Access-Control-Allow-Origin")
			if actualOrigin != tt.expectedOrigin {
				t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, actualOrigin)
			}

			// Check other security headers that should always be present
			expectedMethods := "POST, GET, OPTIONS, PUT, DELETE"
			if actual := rr.Header().Get("Access-Control-Allow-Methods"); actual != expectedMethods {
				t.Errorf("expected Access-Control-Allow-Methods %q, got %q", expectedMethods, actual)
			}

			expectedContentTypeOptions := "nosniff"
			if actual := rr.Header().Get("X-Content-Type-Options"); actual != expectedContentTypeOptions {
				t.Errorf("expected X-Content-Type-Options %q, got %q", expectedContentTypeOptions, actual)
			}
		})
	}
}

func TestSecurityHeadersMiddleware_Options(t *testing.T) {
	// A dummy handler to wrap
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be called for OPTIONS request
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := SecurityHeadersMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Status code should be 200 OK because of the middleware returning early
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
