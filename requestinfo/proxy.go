// Package requestinfo resolves client metadata without replacing RemoteAddr:
// proxy identity authentication must continue to inspect the connection peer.
package requestinfo

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

const maxProxyHops = 32

type Info struct {
	ClientIP string
	HTTPS    bool
}

func parseProxies(entries []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		if addr, err := netip.ParseAddr(entry); err == nil && addr.Zone() == "" {
			addr = addr.Unmap()
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy IP or CIDR: %q", entry)
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return nil, fmt.Errorf("invalid IPv4-mapped trusted proxy CIDR: %q", entry)
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func ValidateTrustedProxies(entries []string) error {
	_, err := parseProxies(entries)
	return err
}

func trusted(addr netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func connectionIP(remote string) (netip.Addr, string) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	addr, err := netip.ParseAddr(host)
	if err != nil || addr.Zone() != "" {
		return netip.Addr{}, host
	}
	addr = addr.Unmap()
	return addr, addr.String()
}

// Resolve ignores forwarded metadata unless the direct TCP peer is explicitly
// trusted. X-Forwarded-For is walked from the nearest proxy towards the client,
// stopping at the first untrusted address; attacker-controlled prefixes to its
// left are never consulted. X-Real-IP is a single-hop fallback only when XFF is
// absent. Malformed authoritative headers fall back to the connection peer.
func Resolve(r *http.Request, entries []string) Info {
	peer, direct := connectionIP(r.RemoteAddr)
	info := Info{ClientIP: direct, HTTPS: r.TLS != nil}
	if !peer.IsValid() || len(entries) == 0 {
		return info
	}
	prefixes, err := parseProxies(entries)
	if err != nil || !trusted(peer, prefixes) {
		return info
	}
	if values := r.Header.Values("X-Forwarded-For"); len(values) != 0 {
		if client, ok := forwardedClient(peer, values, prefixes); ok {
			info.ClientIP = client.String()
		}
	} else if values := r.Header.Values("X-Real-IP"); len(values) == 1 {
		if client, err := netip.ParseAddr(strings.TrimSpace(values[0])); err == nil && client.Zone() == "" {
			info.ClientIP = client.Unmap().String()
		}
	}
	// A single value avoids ambiguous multi-proxy protocol lists. The nearest
	// trusted proxy must overwrite this header with the external scheme.
	if values := r.Header.Values("X-Forwarded-Proto"); len(values) == 1 && strings.EqualFold(strings.TrimSpace(values[0]), "https") {
		info.HTTPS = true
	}
	return info
}

func forwardedClient(peer netip.Addr, values []string, prefixes []netip.Prefix) (netip.Addr, bool) {
	current := peer
	hops := 0
	for i := len(values) - 1; i >= 0; i-- {
		remaining := values[i]
		for {
			if !trusted(current, prefixes) {
				return current, true
			}
			if hops >= maxProxyHops {
				return netip.Addr{}, false
			}
			hops++
			comma := strings.LastIndexByte(remaining, ',')
			entry := remaining
			if comma >= 0 {
				entry = remaining[comma+1:]
			}
			addr, err := netip.ParseAddr(strings.TrimSpace(entry))
			if err != nil || addr.Zone() != "" {
				return netip.Addr{}, false
			}
			current = addr.Unmap()
			if comma < 0 {
				break
			}
			remaining = remaining[:comma]
		}
	}
	return current, true
}
