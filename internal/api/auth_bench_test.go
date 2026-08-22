package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"testing"
)

var (
	benchAdminToken      = "my-super-secret-admin-token-1234567890"
	benchAdminTokenHash  = sha256.Sum256([]byte(benchAdminToken))
	benchAdminTokenBytes = []byte(benchAdminToken)
	benchValidToken      = "my-super-secret-admin-token-1234567890"
)

func BenchmarkAuth_Current(b *testing.B) {
	for i := 0; i < b.N; i++ {
		providedTokenHash := sha256.Sum256([]byte(benchValidToken))
		_ = subtle.ConstantTimeCompare(providedTokenHash[:], benchAdminTokenHash[:]) == 1
	}
}

func BenchmarkAuth_Optimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		providedToken := benchValidToken
		if len(providedToken) == len(benchAdminTokenBytes) {
			_ = subtle.ConstantTimeCompare([]byte(providedToken), benchAdminTokenBytes) == 1
		}
	}
}
