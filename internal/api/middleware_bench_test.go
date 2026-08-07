package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGetClientIP_NoFallback(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.2, 172.16.0.1, 203.0.113.5, 8.8.8.8")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetClientIP(req)
	}
}

func BenchmarkGetClientIP_Fallback(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.2, 172.16.0.1, 192.168.0.5, 10.10.10.10, 172.31.255.255")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetClientIP(req)
	}
}

func BenchmarkRateLimiter_Contention(b *testing.B) {
	rl := NewRateLimiter(10, 10)
	defer rl.Stop()

	// Pre-fill with a large number of IPs to make cleanup slow
	for i := 0; i < 100000; i++ {
		rl.getLimiter(fmt.Sprintf("192.168.1.%d", i))
	}

	b.ResetTimer()

	// Run getLimiter concurrently to simulate traffic
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			rl.getLimiter(fmt.Sprintf("10.0.0.%d", i%1000))
			i++
			if i%100 == 0 {
				rl.cleanup()
			}
		}
	})
}
