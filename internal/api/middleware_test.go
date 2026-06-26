package api

import (
	"net/http"
	"net/http/httptest"
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

func TestSecurityHeadersMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := SecurityHeadersMiddleware(nextHandler)

	t.Run("Sets correct security headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
		if rec.Header().Get("Access-Control-Allow-Methods") != "POST, GET, OPTIONS, PUT, DELETE" {
			t.Errorf("Expected Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE, got %s", rec.Header().Get("Access-Control-Allow-Methods"))
		}
		if rec.Header().Get("Access-Control-Allow-Headers") != "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Token" {
			t.Errorf("Expected Access-Control-Allow-Headers, got %s", rec.Header().Get("Access-Control-Allow-Headers"))
		}
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("Expected X-Content-Type-Options: nosniff, got %s", rec.Header().Get("X-Content-Type-Options"))
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}
	})

	t.Run("OPTIONS method returns 200 early", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		// If nextHandler is called, this would be an error because OPTIONS should return early
		handlerToTest := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler should not be called for OPTIONS request")
		}))

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}
	})
}
