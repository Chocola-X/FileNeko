package requestinfo

import (
	"crypto/tls"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveTrustedProxyMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, remote, xff, real, proto, want string
		proxies                              []string
		secure                               bool
	}{
		{name: "default ignores all headers", remote: "127.0.0.1:1", xff: "198.51.100.2", real: "198.51.100.3", proto: "https", want: "127.0.0.1"},
		{name: "untrusted connection", remote: "198.51.100.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2", proto: "https", want: "198.51.100.1"},
		{name: "trusted nginx", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1/32"}, xff: "198.51.100.2", proto: "https", want: "198.51.100.2", secure: true},
		{name: "real IP fallback", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, real: "192.168.1.2", want: "192.168.1.2"},
		{name: "xff takes precedence", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2", real: "198.51.100.3", want: "198.51.100.2"},
		{name: "ignore forged prefix", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "203.0.113.99, 198.51.100.2", want: "198.51.100.2"},
		{name: "ignore malformed untrusted prefix", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "garbage, unknown, 198.51.100.2", want: "198.51.100.2"},
		{name: "trusted chain", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1", "10.0.0.0/24"}, xff: "203.0.113.99, 198.51.100.2, 10.0.0.2", want: "198.51.100.2"},
		{name: "all hops trusted", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1", "10.0.0.0/24"}, xff: "10.0.0.3, 10.0.0.2", want: "10.0.0.3"},
		{name: "nearest hop untrusted", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2, 10.0.0.2", want: "10.0.0.2"},
		{name: "malformed nearest hop", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2, unknown", real: "198.51.100.3", want: "127.0.0.1"},
		{name: "empty nearest hop", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2, ", want: "127.0.0.1"},
		{name: "invalid real IP", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, real: "198.51.100.2, 198.51.100.3", want: "127.0.0.1"},
		{name: "port is not a forwarded IP", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2:1234", want: "127.0.0.1"},
		{name: "IPv6 peer and client", remote: "[::1]:1", proxies: []string{"::1/128"}, xff: "2001:db8::abcd", proto: "https", want: "2001:db8::abcd", secure: true},
		{name: "IPv4 mapped peer and client", remote: "[::ffff:127.0.0.1]:1", proxies: []string{"127.0.0.1"}, xff: "::ffff:198.51.100.2", want: "198.51.100.2"},
		{name: "mapped trusted CIDR", remote: "127.0.0.1:1", proxies: []string{"::ffff:127.0.0.0/104"}, xff: "198.51.100.2", want: "198.51.100.2"},
		{name: "scoped IPv6 rejected", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, xff: "fe80::1%spoof", want: "127.0.0.1"},
		{name: "Unix socket untrusted", remote: "@", proxies: []string{"127.0.0.1"}, xff: "198.51.100.2", proto: "https", want: "@"},
		{name: "invalid config fails closed", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1", "invalid"}, xff: "198.51.100.2", proto: "https", want: "127.0.0.1"},
		{name: "ambiguous protocol ignored", remote: "127.0.0.1:1", proxies: []string{"127.0.0.1"}, proto: "https, http", want: "127.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tc.remote
			if tc.xff != "" {
				r.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.real != "" {
				r.Header.Set("X-Real-IP", tc.real)
			}
			if tc.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tc.proto)
			}
			r.Header.Set("Forwarded", "for=203.0.113.200;proto=https")
			info := Resolve(r, tc.proxies)
			if info.ClientIP != tc.want || info.HTTPS != tc.secure {
				t.Fatalf("got %+v, want %s secure=%t", info, tc.want, tc.secure)
			}
			if r.RemoteAddr != tc.remote || r.TLS != nil {
				t.Fatal("resolver changed the actual connection metadata")
			}
		})
	}
}

func TestProxyHeaderLinesAndNativeTLS(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:1"
	proxies := []string{"127.0.0.1", "10.0.0.2"}
	r.Header.Add("X-Forwarded-For", "203.0.113.99")
	r.Header.Add("X-Forwarded-For", "198.51.100.2, 10.0.0.2")
	if got := Resolve(r, proxies); got.ClientIP != "198.51.100.2" {
		t.Fatalf("multiple header lines: %+v", got)
	}
	r.Header.Del("X-Forwarded-For")
	r.Header.Add("X-Real-IP", "198.51.100.2")
	r.Header.Add("X-Real-IP", "198.51.100.3")
	r.Header.Add("X-Forwarded-Proto", "https")
	r.Header.Add("X-Forwarded-Proto", "http")
	if got := Resolve(r, proxies); got.ClientIP != "127.0.0.1" || got.HTTPS {
		t.Fatalf("ambiguous single-hop headers accepted: %+v", got)
	}
	r.TLS = &tls.ConnectionState{}
	r.Header.Set("X-Forwarded-Proto", "http")
	if !Resolve(r, proxies).HTTPS {
		t.Fatal("forwarded protocol downgraded native TLS")
	}
	r.Header.Set("X-Forwarded-For", strings.Repeat("10.0.0.2,", maxProxyHops)+"10.0.0.2")
	if got := Resolve(r, proxies); got.ClientIP != "127.0.0.1" {
		t.Fatalf("excessive trusted chain accepted: %+v", got)
	}
}

func TestValidateTrustedProxies(t *testing.T) {
	for _, entries := range [][]string{nil, {}, {"127.0.0.1", "::1"}, {"10.0.0.7/24", "2001:db8::/32"}} {
		if err := ValidateTrustedProxies(entries); err != nil {
			t.Fatal(err)
		}
	}
	for _, entry := range []string{"", "localhost", "127.0.0.1:8080", "invalid", "127.0.0.1/33", "fe80::1%eth0", "::ffff:127.0.0.1/80"} {
		if ValidateTrustedProxies([]string{entry}) == nil {
			t.Fatalf("invalid proxy accepted: %q", entry)
		}
	}
}
