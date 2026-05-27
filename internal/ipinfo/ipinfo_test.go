package ipinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLookup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/json/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"ip":"8.8.8.8",
			"city":"Mountain View",
			"region":"California",
			"country_name":"United States",
			"country_code":"US",
			"latitude":37.4056,
			"longitude":-122.0775,
			"timezone":"America/Los_Angeles",
			"asn":"AS15169",
			"org":"Google LLC"
		}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	client.Provider = "test"
	info := client.Lookup(context.Background(), "8.8.8.8")
	if info.Error != "" {
		t.Fatalf("unexpected error: %s", info.Error)
	}
	if info.City != "Mountain View" || info.ASN != "AS15169" || info.ISP != "Google LLC" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestLookupSkipsPrivateIP(t *testing.T) {
	client := NewClient()
	info := client.Lookup(context.Background(), "192.168.1.1")
	if info.Error == "" {
		t.Fatal("expected private IP error")
	}
}

func TestLookupManyDeduplicates(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"ip":"8.8.8.8","country_name":"United States","org":"Google LLC"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	client.Timeout = time.Second
	infos := client.LookupMany(context.Background(), []string{"8.8.8.8", "8.8.8.8"})
	if len(infos) != 1 {
		t.Fatalf("unexpected count: %d", len(infos))
	}
	if hits != 1 {
		t.Fatalf("expected one HTTP hit, got %d", hits)
	}
}
