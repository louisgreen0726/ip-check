package resolver

import "testing"

func TestQueryType(t *testing.T) {
	for _, name := range []string{"A", "aaaa", "HTTPS", "SVCB", "TXT"} {
		if _, err := QueryType(name); err != nil {
			t.Fatalf("QueryType(%q) returned error: %v", name, err)
		}
	}
}

func TestNewQueryUsesFQDNAndEDNS(t *testing.T) {
	msg := NewQuery("example.com", 1, Options{EDNS: true, UDPSize: 1232, DNSSEC: true})
	if got := msg.Question[0].Name; got != "example.com." {
		t.Fatalf("question name = %q", got)
	}
	if opt := msg.IsEdns0(); opt == nil {
		t.Fatal("expected EDNS0 option")
	} else if opt.UDPSize() != 1232 {
		t.Fatalf("unexpected UDP size: %d", opt.UDPSize())
	}
}
