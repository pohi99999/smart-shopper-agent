package api

import (
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimiter struct holds the limiters for different IP addresses
type rateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter (10 requests per minute = ~0.16 req/sec)
func NewRateLimiter(r rate.Limit, b int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter returns the limiter for the provided IP address.
func (i *rateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// Cleanup periodically removes limiters for IPs that haven't been seen in a while (omitted for brevity, or we can use a simpler approach)
// Since this is a simple implementation, we'll just keep them. In a real scenario, use a cache with TTL.

var limiter = NewRateLimiter(rate.Every(time.Minute/10), 10) // 10 requests per minute

// RateLimitMiddleware applies rate limiting per IP address
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract IP from RemoteAddr
		ip := r.RemoteAddr
		// Could also handle X-Forwarded-For if behind a proxy
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		limiter := limiter.getLimiter(ip)
		if !limiter.Allow() {
			SendJSONError(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// SecurityHeadersMiddleware adds basic security HTTP headers to responses
func SecurityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Token")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}
