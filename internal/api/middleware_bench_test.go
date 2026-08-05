package api

import (
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
