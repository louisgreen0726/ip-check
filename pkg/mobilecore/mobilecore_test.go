package mobilecore

import (
	"encoding/json"
	"testing"
)

func TestResolveJSONInvalidDomain(t *testing.T) {
	raw := ResolveJSON(`{"domainText":"a..b.com","dns":["udp://127.0.0.1:1"],"types":["A"],"timeoutMs":50,"retries":0}`)
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if len(resp.Results) != 1 || resp.Results[0].Status != "INVALID_DOMAIN" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestResolveJSONBadRequest(t *testing.T) {
	raw := ResolveJSON(`{`)
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp.Error == "" {
		t.Fatal("expected error")
	}
}

func TestHealthJSON(t *testing.T) {
	raw := HealthJSON()
	var resp map[string]string
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["status"] != "ok" || resp["version"] == "" {
		t.Fatalf("unexpected health: %+v", resp)
	}
}
