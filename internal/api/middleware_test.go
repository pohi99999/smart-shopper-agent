package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
			expectedOrigin: "*",
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
