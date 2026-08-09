package api

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// visitor struct holds the rate limiter and the last time the IP was seen
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

const numShards = 32

type rateLimiterShard struct {
	visitors map[string]*visitor
	mu       sync.Mutex
}

// rateLimiter struct holds the limiters for different IP addresses
type rateLimiter struct {
	shards [numShards]*rateLimiterShard
	rate   rate.Limit
	burst  int
	stop   chan struct{}
}

func getShardIndex(ip string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(ip); i++ {
		hash ^= uint32(ip[i])
		hash *= 16777619
	}
	return hash % numShards
}

// NewRateLimiter creates a new rate limiter (10 requests per minute = ~0.16 req/sec)
func NewRateLimiter(r rate.Limit, b int) *rateLimiter {
	rl := &rateLimiter{
		rate:  r,
		burst: b,
		stop:  make(chan struct{}),
	}
	for i := 0; i < numShards; i++ {
		rl.shards[i] = &rateLimiterShard{
			visitors: make(map[string]*visitor),
		}
	}
	go rl.runCleanup()
	return rl
}

func (i *rateLimiter) runCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			i.cleanup()
		case <-i.stop:
			return
		}
	}
}

// Stop terminates the background cleanup goroutine
func (i *rateLimiter) Stop() {
	close(i.stop)
}

// getLimiter returns the limiter for the provided IP address
func (i *rateLimiter) getLimiter(ip string) *rate.Limiter {
	shard := i.shards[getShardIndex(ip)]
	shard.mu.Lock()
	now := time.Now()

	v, exists := shard.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		v = &visitor{limiter: limiter, lastSeen: now}
		shard.visitors[ip] = v
	} else {
		v.lastSeen = now
	}
	shard.mu.Unlock()

	return v.limiter
}

// cleanup removes old entries from the visitors map.
func (i *rateLimiter) cleanup() {
	now := time.Now()
	for j := 0; j < numShards; j++ {
		shard := i.shards[j]
		shard.mu.Lock()
		for ip, v := range shard.visitors {
			if now.Sub(v.lastSeen) > 3*time.Minute {
				delete(shard.visitors, ip)
			}
		}
		shard.mu.Unlock()
	}
}

var limiter = NewRateLimiter(rate.Every(time.Minute/10), 10) // 10 requests per minute

var (
	trustedProxiesCache atomic.Value // Stores []*net.IPNet
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

	trustedProxiesCache.Store(newCache)
}

func isTrustedProxy(ip string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	cache, ok := trustedProxiesCache.Load().([]*net.IPNet)
	if !ok {
		return false
	}

	for _, ipNet := range cache {
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
		var firstValid string
		for i := len(ips) - 1; i >= 0; i-- {
			part := strings.TrimSpace(ips[i])
			parsed := net.ParseIP(part)
			if parsed != nil {
				if !parsed.IsPrivate() && !parsed.IsLoopback() {
					return part
				}
				firstValid = part
			}
		}
		if firstValid != "" {
			return firstValid
		}
	}
	return ip
}

// RateLimitMiddleware applies rate limiting per IP address
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		l := limiter.getLimiter(ip)
		if !l.Allow() {
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
