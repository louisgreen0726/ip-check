package resolver

import (
	"context"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/miekg/dns"

	"ipcheck/internal/endpoint"
)

func TestExchangeUDPCustomPort(t *testing.T) {
	addr, shutdown := startDNSServer(t, "udp")
	defer shutdown()

	ep := endpoint.Endpoint{Scheme: endpoint.UDP, Host: "127.0.0.1", Port: addr.Port, Raw: "test-udp"}
	msg := NewQuery("example.com.", dns.TypeA, Options{EDNS: true, UDPSize: 1232})
	result, err := Exchange(context.Background(), ep, msg, Options{Timeout: time.Second, EDNS: true, UDPSize: 1232})
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}
	parsed := ParseMessage(result.Message)
	if len(parsed.IPs) != 1 || parsed.IPs[0] != "192.0.2.42" {
		t.Fatalf("unexpected IPs: %#v", parsed.IPs)
	}
}

func TestExchangeTCPCustomPort(t *testing.T) {
	addr, shutdown := startDNSServer(t, "tcp")
	defer shutdown()

	ep := endpoint.Endpoint{Scheme: endpoint.TCP, Host: "127.0.0.1", Port: addr.Port, Raw: "test-tcp"}
	msg := NewQuery("example.com.", dns.TypeA, Options{EDNS: true, UDPSize: 1232})
	result, err := Exchange(context.Background(), ep, msg, Options{Timeout: time.Second, EDNS: true, UDPSize: 1232})
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}
	parsed := ParseMessage(result.Message)
	if len(parsed.IPs) != 1 || parsed.IPs[0] != "192.0.2.42" {
		t.Fatalf("unexpected IPs: %#v", parsed.IPs)
	}
}

func TestExchangeDoHCustomPort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wire []byte
		switch r.Method {
		case http.MethodPost:
			var err error
			wire, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
		case http.MethodGet:
			var err error
			wire, err = base64.RawURLEncoding.DecodeString(r.URL.Query().Get("dns"))
			if err != nil {
				t.Fatalf("decode dns query: %v", err)
			}
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}

		req := new(dns.Msg)
		if err := req.Unpack(wire); err != nil {
			t.Fatalf("unpack query: %v", err)
		}
		resp := replyFor(req)
		data, err := resp.Pack()
		if err != nil {
			t.Fatalf("pack response: %v", err)
		}
		w.Header().Set("Content-Type", "application/dns-message")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	portNum, err := net.LookupPort("tcp", port)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	ep := endpoint.Endpoint{Scheme: endpoint.HTTP, Host: host, Port: portNum, Path: "/dns-query", Raw: server.URL}
	msg := NewQuery("example.com.", dns.TypeA, Options{EDNS: true, UDPSize: 1232})
	result, err := Exchange(context.Background(), ep, msg, Options{Timeout: time.Second, EDNS: true, UDPSize: 1232, DoHMethod: "POST"})
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}
	parsed := ParseMessage(result.Message)
	if len(parsed.IPs) != 1 || parsed.IPs[0] != "192.0.2.42" {
		t.Fatalf("unexpected IPs: %#v", parsed.IPs)
	}
}

func startDNSServer(t *testing.T, network string) (*net.TCPAddr, func()) {
	t.Helper()

	handler := dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		resp := replyFor(req)
		if err := w.WriteMsg(resp); err != nil {
			t.Errorf("WriteMsg failed: %v", err)
		}
	})

	switch network {
	case "udp":
		pc, err := net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen udp: %v", err)
		}
		addr := pc.LocalAddr().(*net.UDPAddr)
		server := &dns.Server{PacketConn: pc, Handler: handler}
		go func() {
			if err := server.ActivateAndServe(); err != nil {
				t.Logf("udp dns server stopped: %v", err)
			}
		}()
		return &net.TCPAddr{IP: addr.IP, Port: addr.Port}, func() { _ = server.Shutdown() }
	case "tcp":
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen tcp: %v", err)
		}
		addr := listener.Addr().(*net.TCPAddr)
		server := &dns.Server{Listener: listener, Net: "tcp", Handler: handler}
		go func() {
			if err := server.ActivateAndServe(); err != nil {
				t.Logf("tcp dns server stopped: %v", err)
			}
		}()
		return addr, func() { _ = server.Shutdown() }
	default:
		t.Fatalf("unsupported network: %s", network)
		return nil, nil
	}
}

func replyFor(req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	resp.SetReply(req)
	if len(req.Question) == 0 {
		return resp
	}
	q := req.Question[0]
	if q.Qtype == dns.TypeA {
		resp.Answer = append(resp.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A:   net.ParseIP("192.0.2.42"),
		})
	}
	return resp
}
