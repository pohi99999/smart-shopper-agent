package api

import (
	"net"
)

type trieNode struct {
	children [2]*trieNode
	isTerminal bool
}

type CIDRTrie struct {
	v4    *trieNode
	v6    *trieNode
	cidrs []*net.IPNet // retained for backward-compatibility with tests
}

func NewCIDRTrie() *CIDRTrie {
	return &CIDRTrie{
		v4:    &trieNode{},
		v6:    &trieNode{},
		cidrs: []*net.IPNet{},
	}
}

func (t *CIDRTrie) AddIPNet(ipNet *net.IPNet) {
	t.cidrs = append(t.cidrs, ipNet)

	isV4 := ipNet.IP.To4() != nil

	var ipBytes []byte
	if isV4 {
		ipBytes = ipNet.IP.To4()
	} else {
		ipBytes = ipNet.IP.To16()
	}

	ones, bits := ipNet.Mask.Size()
	if bits == 128 && isV4 {
		// Adjust the ones count for IPv4 which only has 32 bits,
		// but was given a 128-bit mask.
		// A /128 mask on an IPv4-mapped IPv6 means /32 in IPv4 land.
		if ones == 128 {
			ones = 32
		}
	}

	node := t.v4
	if !isV4 {
		node = t.v6
	}

	for i := 0; i < ones; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> bitIndex) & 1

		if node.children[bit] == nil {
			node.children[bit] = &trieNode{}
		}
		node = node.children[bit]
	}
	node.isTerminal = true
}

func (t *CIDRTrie) Contains(ip net.IP) bool {
	isV4 := ip.To4() != nil

	var ipBytes []byte
	if isV4 {
		ipBytes = ip.To4()
	} else {
		ipBytes = ip.To16()
	}


	node := t.v4
	if !isV4 {
		node = t.v6
	}

	if node.isTerminal {
		return true
	}

	for i := 0; i < len(ipBytes)*8; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		bit := (ipBytes[byteIndex] >> bitIndex) & 1

		node = node.children[bit]
		if node == nil {
			return false
		}
		if node.isTerminal {
			return true
		}
	}

	return false
}
