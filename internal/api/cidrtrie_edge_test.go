package api

import (
	"net"
	"testing"
)

func TestCIDRTrie_IPv4MappedIPv6(t *testing.T) {
	trie := NewCIDRTrie()

	// Create an IPv4-mapped IPv6 address with a 128-bit mask
	ip := net.ParseIP("::ffff:192.168.1.1")
	mask := net.CIDRMask(128, 128)
	ipNet := &net.IPNet{IP: ip, Mask: mask}

	// This should not panic
	trie.AddIPNet(ipNet)

	if !trie.Contains(ip) {
		t.Errorf("expected trie to contain %s", ip)
	}
}
