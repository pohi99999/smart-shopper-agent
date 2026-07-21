package api

import (
	"net"

	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor struct holds the rate limiter and the last time the IP was seen
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiter struct holds the limiters for different IP addresses
type rateLimiter struct {
	visitors    map[string]*visitor
	mu          sync.Mutex
	rate        rate.Limit
	burst       int
	lastCleanup time.Time
}

// NewRateLimiter creates a new rate limiter (10 requests per minute = ~0.16 req/sec)
func NewRateLimiter(r rate.Limit, b int) *rateLimiter {
	return &rateLimiter{
		visitors:    make(map[string]*visitor),
		rate:        r,
		burst:       b,
		lastCleanup: time.Now(),
	}
}

// getLimiter returns the limiter for the provided IP address and performs opportunistic cleanup.
func (i *rateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()

	now := time.Now()
	shouldCleanup := false

	// Opportunistic cleanup: Every 3 minutes, remove entries that haven't been seen for 3 minutes
	if now.Sub(i.lastCleanup) > 3*time.Minute {
		shouldCleanup = true
		i.lastCleanup = now
	}

	v, exists := i.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		v = &visitor{limiter: limiter, lastSeen: now}
		i.visitors[ip] = v
	} else {
		v.lastSeen = now
	}

	i.mu.Unlock()

	if shouldCleanup {
		go i.cleanup()
	}

	return v.limiter
}

// cleanup removes old entries from the visitors map.
func (i *rateLimiter) cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()
	now := time.Now()
	for ip, v := range i.visitors {
		if now.Sub(v.lastSeen) > 3*time.Minute {
			delete(i.visitors, ip)
		}
	}
}

var limiter = NewRateLimiter(rate.Every(time.Minute/10), 10) // 10 requests per minute

var (
	trustedProxiesCache []*net.IPNet
	trustedProxiesMu    sync.RWMutex
)

func init() {
	LoadTrustedProxies()
}

// LoadTrustedProxies reads the TRUSTED_PROXIES environment variable and updates the internal cache.
func LoadTrustedProxies() {
	trustedProxiesEnv := os.Getenv("TRUSTED_PROXIES")
	var newCache []*net.IPNet

	if trustedProxiesEnv != "" {
		proxies := strings.Split(trustedProxiesEnv, ",")
		for _, proxy := range proxies {
			proxy = strings.TrimSpace(proxy)
			if proxy == "" {
				continue
			}

			if strings.Contains(proxy, "/") {
				_, ipNet, err := net.ParseCIDR(proxy)
				if err == nil {
					newCache = append(newCache, ipNet)
				}
			} else {
				trustedIP := net.ParseIP(proxy)
				if trustedIP != nil {
					// Convert single IP to a /32 or /128 CIDR network for uniform matching
					var mask net.IPMask
					if trustedIP.To4() != nil {
						mask = net.CIDRMask(32, 32)
					} else {
						mask = net.CIDRMask(128, 128)
					}
					ipNet := &net.IPNet{IP: trustedIP, Mask: mask}
					newCache = append(newCache, ipNet)
				}
			}
		}
	}

	trustedProxiesMu.Lock()
	trustedProxiesCache = newCache
	trustedProxiesMu.Unlock()
}

func isTrustedProxy(ip string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	trustedProxiesMu.RLock()
	defer trustedProxiesMu.RUnlock()

	for _, ipNet := range trustedProxiesCache {
		if ipNet.Contains(clientIP) {
			return true
		}
	}
	return false
}

// GetClientIP extracts the real client IP address securely, handling X-Forwarded-For.
func GetClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}

	if !isTrustedProxy(ip) {
		return ip
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		for i := len(ips) - 1; i >= 0; i-- {
			part := strings.TrimSpace(ips[i])
			parsed := net.ParseIP(part)
			if parsed != nil {
				if parsed.IsPrivate() || parsed.IsLoopback() {
					continue
				}
				return part
			}
		}
		for _, ipStr := range ips {
			part := strings.TrimSpace(ipStr)
			parsed := net.ParseIP(part)
			if parsed != nil {
				return part
			}
		}
	}
	return ip
}

// RateLimitMiddleware applies rate limiting per IP address
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
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
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	return func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
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
