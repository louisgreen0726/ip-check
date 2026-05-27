package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/miekg/dns"
)

func TestServerAssetsAndHealth(t *testing.T) {
	mux := newServerMux()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("IP Check")) {
		t.Fatal("index did not contain app title")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/health status = %d", rec.Code)
	}
}

func TestResolveAPIInvalidDomain(t *testing.T) {
	mux := newServerMux()
	body := bytes.NewBufferString(`{"domainText":"a..b.com","dns":["udp://127.0.0.1:1"],"types":["A"],"timeoutMs":50,"retries":0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/resolve", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/resolve status = %d body=%s", rec.Code, rec.Body.String())
	}

	var resp resolveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Summary.Invalid != 1 || len(resp.Results) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Results[0].Status != "INVALID_DOMAIN" {
		t.Fatalf("unexpected status: %s", resp.Results[0].Status)
	}
}

func TestResolveAPISuccessOnCustomUDPPort(t *testing.T) {
	addr, shutdown := startAPIDNSTestServer(t)
	defer shutdown()

	payload := map[string]any{
		"domainText": "example.com",
		"dns":        []string{"udp://" + addr.String()},
		"types":      []string{"A"},
		"timeoutMs":  1000,
		"retries":    0,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/resolve", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	newServerMux().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/resolve status = %d body=%s", rec.Code, rec.Body.String())
	}

	var resp resolveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Summary.OK != 1 || len(resp.Results) != 1 {
		t.Fatalf("unexpected summary/results: %+v", resp)
	}
	if len(resp.Results[0].IPs) != 1 || resp.Results[0].IPs[0] != "192.0.2.55" {
		t.Fatalf("unexpected IPs: %#v", resp.Results[0].IPs)
	}
}

func TestResolveAPIBadRequest(t *testing.T) {
	tests := []string{
		`{`,
		`{"domainText":"","dns":["udp://127.0.0.1:1"],"types":["A"]}`,
		`{"domainText":"example.com","dns":["udp://127.0.0.1:70000"],"types":["A"]}`,
		`{"domainText":"example.com","dns":["udp://127.0.0.1:53"],"types":["NOT_A_TYPE"]}`,
	}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPost, "/api/resolve", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		newServerMux().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s status = %d, want 400", body, rec.Code)
		}
	}
}

func startAPIDNSTestServer(t *testing.T) (*net.UDPAddr, func()) {
	t.Helper()

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	addr := pc.LocalAddr().(*net.UDPAddr)
	server := &dns.Server{
		PacketConn: pc,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
			resp := new(dns.Msg)
			resp.SetReply(req)
			if len(req.Question) > 0 && req.Question[0].Qtype == dns.TypeA {
				resp.Answer = append(resp.Answer, &dns.A{
					Hdr: dns.RR_Header{Name: req.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP("192.0.2.55"),
				})
			}
			if err := w.WriteMsg(resp); err != nil {
				t.Errorf("WriteMsg failed: %v", err)
			}
		}),
	}
	go func() {
		if err := server.ActivateAndServe(); err != nil {
			t.Logf("dns test server stopped: %v", err)
		}
	}()
	return addr, func() { _ = server.Shutdown() }
}
