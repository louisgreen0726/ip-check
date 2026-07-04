package endpoint

import "testing"

func TestParseCustomPorts(t *testing.T) {
	tests := []struct {
		raw    string
		scheme Scheme
		host   string
		port   int
		path   string
	}{
		{"udp://8.8.8.8:5353", UDP, "8.8.8.8", 5353, ""},
		{"tcp://1.1.1.1:9953", TCP, "1.1.1.1", 9953, ""},
		{"tls://dns.example.com:1853", TLS, "dns.example.com", 1853, ""},
		{"dot://dns.example.com:2853", TLS, "dns.example.com", 2853, ""},
		{"https://dns.example.com:8443/dns-query", HTTPS, "dns.example.com", 8443, "/dns-query"},
		{"doh://dns.example.com:9443/custom", HTTPS, "dns.example.com", 9443, "/custom"},
		{"quic://dns.example.com:8853", QUIC, "dns.example.com", 8853, ""},
		{"doq://dns.example.com:8854", QUIC, "dns.example.com", 8854, ""},
	}

	for _, tt := range tests {
		ep, err := Parse(tt.raw)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", tt.raw, err)
		}
		if ep.Scheme != tt.scheme || ep.Host != tt.host || ep.Port != tt.port || ep.Path != tt.path {
			t.Fatalf("Parse(%q) = %+v", tt.raw, ep)
		}
	}
}

func TestParseDefaultPorts(t *testing.T) {
	tests := []struct {
		raw  string
		port int
	}{
		{"8.8.8.8", 53},
		{"udp://8.8.8.8", 53},
		{"tcp://8.8.8.8", 53},
		{"tls://1.1.1.1", 853},
		{"https://dns.google/dns-query", 443},
		{"quic://dns.adguard-dns.com", 853},
	}

	for _, tt := range tests {
		ep, err := Parse(tt.raw)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", tt.raw, err)
		}
		if ep.Port != tt.port {
			t.Fatalf("Parse(%q) port = %d, want %d", tt.raw, ep.Port, tt.port)
		}
	}
}

func TestParseIPv6CustomPort(t *testing.T) {
	ep, err := Parse("udp://[2606:4700:4700::1111]:5353")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if ep.Host != "2606:4700:4700::1111" || ep.Port != 5353 {
		t.Fatalf("unexpected IPv6 endpoint: %+v", ep)
	}
}

func TestRejectInvalidPorts(t *testing.T) {
	for _, raw := range []string{
		"udp://8.8.8.8:0",
		"udp://8.8.8.8:65536",
		"udp://8.8.8.8:abc",
	} {
		if _, err := Parse(raw); err == nil {
			t.Fatalf("expected error for %s", raw)
		}
	}
}

func TestRejectHostWhitespace(t *testing.T) {
	for _, raw := range []string{
		"bad host",
		"udp://bad%20host:53",
	} {
		if _, err := Parse(raw); err == nil {
			t.Fatalf("expected error for %s", raw)
		}
	}
}

func TestEndpointTLSName(t *testing.T) {
	ep, err := Parse("tls://1.1.1.1:853?sni=cloudflare-dns.com")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if ep.TLSName() != "cloudflare-dns.com" {
		t.Fatalf("unexpected TLSName: %s", ep.TLSName())
	}
}
