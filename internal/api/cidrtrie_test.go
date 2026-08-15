package api

import (
	"fmt"
	"net"
	"testing"
)

func TestCIDRTrie_Basic(t *testing.T) {
	trie := NewCIDRTrie()

	_, ipNet1, _ := net.ParseCIDR("192.168.1.0/24")
	trie.AddIPNet(ipNet1)

	_, ipNet2, _ := net.ParseCIDR("10.0.0.0/8")
	trie.AddIPNet(ipNet2)

	_, ipNet3, _ := net.ParseCIDR("2001:db8::/32")
	trie.AddIPNet(ipNet3)

	tests := []struct{
		ip string
		expected bool
	}{
		{"192.168.1.100", true},
		{"192.168.2.100", false},
		{"10.5.5.5", true},
		{"11.5.5.5", false},
		{"2001:db8::1", true},
		{"2001:db9::1", false},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if trie.Contains(ip) != tt.expected {
			t.Errorf("expected %v for %s, got %v", tt.expected, tt.ip, !tt.expected)
		}
	}
}

func BenchmarkCIDRTrieContains(b *testing.B) {
	trie := NewCIDRTrie()
	for i := 0; i < 100; i++ {
		_, ipNet, _ := net.ParseCIDR(fmt.Sprintf("10.%d.0.0/16", i))
		trie.AddIPNet(ipNet)
	}
	_, ipNet, _ := net.ParseCIDR("192.168.1.100/32")
	trie.AddIPNet(ipNet)

	ip := net.ParseIP("192.168.1.100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Contains(ip)
	}
}

func BenchmarkCIDRTrieContains_Miss(b *testing.B) {
	trie := NewCIDRTrie()
	for i := 0; i < 100; i++ {
		_, ipNet, _ := net.ParseCIDR(fmt.Sprintf("10.%d.0.0/16", i))
		trie.AddIPNet(ipNet)
	}

	ip := net.ParseIP("192.168.1.100")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Contains(ip)
	}
}
