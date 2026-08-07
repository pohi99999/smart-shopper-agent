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
		limiter.Stop() // Stop the newly created global one for tests
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
		limiter.Stop() // Stop the newly created global one for tests
		limiter = originalLimiter
	}()

	originalProxies := os.Getenv("TRUSTED_PROXIES")
	os.Setenv("TRUSTED_PROXIES", "192.168.1.1, 192.168.1.2")
	LoadTrustedProxies()
	defer func() {
		os.Setenv("TRUSTED_PROXIES", originalProxies)
		LoadTrustedProxies()
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
	originalProxies := os.Getenv("TRUSTED_PROXIES")
	os.Setenv("TRUSTED_PROXIES", "10.0.0.1, 192.168.1.100")
	LoadTrustedProxies()
	defer func() {
		os.Setenv("TRUSTED_PROXIES", originalProxies)
		LoadTrustedProxies()
	}()

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
		{"Untrusted Proxy", "10.0.0.2:1234", "8.8.8.8", "10.0.0.2"},
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

func TestLoadTrustedProxies(t *testing.T) {
	// Save original env var to restore later
	originalEnv := os.Getenv("TRUSTED_PROXIES")
	defer func() {
		os.Setenv("TRUSTED_PROXIES", originalEnv)
		LoadTrustedProxies() // Restore original state
	}()

	tests := []struct {
		name        string
		envVar      string
		expectedLen int
		expected    []string // String representations of expected CIDRs
	}{
		{
			name:        "Empty env var",
			envVar:      "",
			expectedLen: 0,
			expected:    nil,
		},
		{
			name:        "Single IPv4",
			envVar:      "192.168.1.1",
			expectedLen: 1,
			expected:    []string{"192.168.1.1/32"},
		},
		{
			name:        "Single IPv6",
			envVar:      "2001:db8::1",
			expectedLen: 1,
			expected:    []string{"2001:db8::1/128"},
		},
		{
			name:        "IPv4 CIDR",
			envVar:      "10.0.0.0/8",
			expectedLen: 1,
			expected:    []string{"10.0.0.0/8"},
		},
		{
			name:        "IPv6 CIDR",
			envVar:      "2001:db8::/32",
			expectedLen: 1,
			expected:    []string{"2001:db8::/32"},
		},
		{
			name:        "Multiple mixed values with spaces",
			envVar:      "192.168.1.1, 10.0.0.0/8 , 2001:db8::1",
			expectedLen: 3,
			expected:    []string{"192.168.1.1/32", "10.0.0.0/8", "2001:db8::1/128"},
		},
		{
			name:        "Invalid IPs and CIDRs",
			envVar:      "invalid-ip, 192.168.1.999, 10.0.0.0/99, 192.168.1.1",
			expectedLen: 1,
			expected:    []string{"192.168.1.1/32"},
		},
		{
			name:        "Empty proxies between commas",
			envVar:      "192.168.1.1,, ,10.0.0.0/8",
			expectedLen: 2,
			expected:    []string{"192.168.1.1/32", "10.0.0.0/8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TRUSTED_PROXIES", tt.envVar)
			LoadTrustedProxies()

			trustedProxiesMu.RLock()
			defer trustedProxiesMu.RUnlock()

			if len(trustedProxiesCache) != tt.expectedLen {
				t.Errorf("Expected cache length %d, got %d", tt.expectedLen, len(trustedProxiesCache))
			}

			if tt.expectedLen > 0 {
				for i, expectedCIDR := range tt.expected {
					if trustedProxiesCache[i].String() != expectedCIDR {
						t.Errorf("Expected CIDR at index %d to be %s, got %s", i, expectedCIDR, trustedProxiesCache[i].String())
					}
				}
			}
		})
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	// Create a new rate limiter
	rl := NewRateLimiter(rate.Every(time.Minute/10), 1)
	// Stop the background cleanup routine immediately so we can test it manually
	rl.Stop()

	now := time.Now()

	// Helper func to add a visitor
	addVisitor := func(ip string, offset time.Duration) {
		shard := rl.shards[getShardIndex(ip)]
		shard.mu.Lock()
		defer shard.mu.Unlock()
		shard.visitors[ip] = &visitor{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: now.Add(offset),
		}
	}

	// Helper func to check if visitor exists
	checkVisitor := func(ip string) bool {
		shard := rl.shards[getShardIndex(ip)]
		shard.mu.Lock()
		defer shard.mu.Unlock()
		_, exists := shard.visitors[ip]
		return exists
	}

	// Manually populate the visitors map
	addVisitor("recent", 0)
	addVisitor("old", -4*time.Minute)
	addVisitor("boundary", -2*time.Minute)

	// Run cleanup manually
	rl.cleanup()

	// Verify results
	if !checkVisitor("recent") {
		t.Errorf("Expected 'recent' visitor to still exist, but it was deleted")
	}

	if !checkVisitor("boundary") {
		t.Errorf("Expected 'boundary' visitor to still exist, but it was deleted")
	}

	if checkVisitor("old") {
		t.Errorf("Expected 'old' visitor to be deleted, but it still exists")
	}
}

func TestIsTrustedProxy(t *testing.T) {
	// Save original env var to restore later
	originalEnv := os.Getenv("TRUSTED_PROXIES")
	defer func() {
		os.Setenv("TRUSTED_PROXIES", originalEnv)
		LoadTrustedProxies() // Restore original state
	}()

	// Setup trusted proxies for testing
	os.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 192.168.1.100/32, 2001:db8::/32")
	LoadTrustedProxies()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IP within 10.0.0.0/8 subnet",
			ip:       "10.1.2.3",
			expected: true,
		},
		{
			name:     "Exact IP match 192.168.1.100/32",
			ip:       "192.168.1.100",
			expected: true,
		},
		{
			name:     "IPv6 within subnet",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "IP outside trusted subnets",
			ip:       "192.168.1.101",
			expected: false,
		},
		{
			name:     "Public IP outside trusted subnets",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "Invalid IP string",
			ip:       "invalid-ip",
			expected: false,
		},
		{
			name:     "Empty IP string",
			ip:       "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrustedProxy(tt.ip)
			if result != tt.expected {
				t.Errorf("isTrustedProxy(%q) = %v; want %v", tt.ip, result, tt.expected)
			}
		})
	}
}
