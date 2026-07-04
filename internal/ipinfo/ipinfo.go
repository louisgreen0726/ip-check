package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Info struct {
	IP          string  `json:"ip"`
	Provider    string  `json:"provider"`
	Country     string  `json:"country,omitempty"`
	CountryCode string  `json:"countryCode,omitempty"`
	Region      string  `json:"region,omitempty"`
	City        string  `json:"city,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	ASN         string  `json:"asn,omitempty"`
	Org         string  `json:"org,omitempty"`
	ISP         string  `json:"isp,omitempty"`
	Error       string  `json:"error,omitempty"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
	Provider   string

	mu    sync.Mutex
	cache map[string]Info
}

func NewClient() *Client {
	return &Client{
		BaseURL:  "https://ipapi.co",
		Timeout:  1800 * time.Millisecond,
		Provider: "ipapi.co",
		cache:    map[string]Info{},
	}
}

func (c *Client) LookupMany(ctx context.Context, ips []string) map[string]Info {
	out := make(map[string]Info, len(ips))
	unique := uniqueIPs(ips)
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 6)

	for _, ip := range unique {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			info := c.Lookup(ctx, ip)
			mu.Lock()
			out[ip] = info
			mu.Unlock()
		}(ip)
	}
	wg.Wait()
	return out
}

func (c *Client) Lookup(ctx context.Context, ip string) Info {
	ip = strings.TrimSpace(ip)
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return Info{IP: ip, Provider: c.provider(), Error: "invalid IP"}
	}
	if reason := specialUseIPReason(ip); reason != "" {
		return Info{IP: ip, Provider: c.provider(), Error: reason}
	}

	if cached, ok := c.getCached(ip); ok {
		return cached
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 1800 * time.Millisecond
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	endpoint, err := url.JoinPath(c.baseURL(), ip, "json")
	if err != nil {
		return Info{IP: ip, Provider: c.provider(), Error: err.Error()}
	}
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint+"/", nil)
	if err != nil {
		return Info{IP: ip, Provider: c.provider(), Error: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ipcheck/0.1.2")

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return c.save(ip, Info{IP: ip, Provider: c.provider(), Error: err.Error()})
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return c.save(ip, Info{IP: ip, Provider: c.provider(), Error: fmt.Sprintf("HTTP %d", resp.StatusCode)})
	}

	var payload ipapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return c.save(ip, Info{IP: ip, Provider: c.provider(), Error: err.Error()})
	}
	if payload.Error {
		reason := payload.Reason
		if reason == "" {
			reason = "lookup failed"
		}
		return c.save(ip, Info{IP: ip, Provider: c.provider(), Error: reason})
	}

	info := Info{
		IP:          firstNonEmpty(payload.IP, ip),
		Provider:    c.provider(),
		Country:     payload.CountryName,
		CountryCode: payload.CountryCode,
		Region:      payload.Region,
		City:        payload.City,
		Timezone:    payload.Timezone,
		Latitude:    payload.Latitude,
		Longitude:   payload.Longitude,
		ASN:         payload.ASN,
		Org:         payload.Org,
		ISP:         payload.Org,
	}
	return c.save(ip, info)
}

type ipapiResponse struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	CountryName string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	ASN         string  `json:"asn"`
	Org         string  `json:"org"`
	Error       bool    `json:"error"`
	Reason      string  `json:"reason"`
}

func (c *Client) baseURL() string {
	if c != nil && strings.TrimSpace(c.BaseURL) != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return "https://ipapi.co"
}

func (c *Client) provider() string {
	if c != nil && strings.TrimSpace(c.Provider) != "" {
		return c.Provider
	}
	return "ipapi.co"
}

func (c *Client) getCached(ip string) (Info, bool) {
	if c == nil {
		return Info{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = map[string]Info{}
	}
	info, ok := c.cache[ip]
	return info, ok
}

func (c *Client) save(ip string, info Info) Info {
	if c == nil {
		return info
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = map[string]Info{}
	}
	c.cache[ip] = info
	return info
}

func uniqueIPs(ips []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		unique = append(unique, ip)
	}
	return unique
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func specialUseIPReason(raw string) string {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil || !addr.IsValid() {
		return "invalid IP"
	}
	if addr.IsLoopback() || addr.IsUnspecified() || addr.IsMulticast() || addr.IsLinkLocalUnicast() {
		return "private or non-routable IP"
	}
	if addr.IsPrivate() {
		return "private or non-routable IP"
	}
	if addr.Is4() && specialIPv4Prefix(addr) {
		return "special-use IP"
	}
	if addr.Is6() && specialIPv6Prefix(addr) {
		return "special-use IP"
	}
	return ""
}

func specialIPv4Prefix(addr netip.Addr) bool {
	prefixes := []string{
		"0.0.0.0/8",
		"100.64.0.0/10",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
	}
	for _, raw := range prefixes {
		prefix, err := netip.ParsePrefix(raw)
		if err == nil && prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func specialIPv6Prefix(addr netip.Addr) bool {
	prefixes := []string{
		"2001:db8::/32",
	}
	for _, raw := range prefixes {
		prefix, err := netip.ParsePrefix(raw)
		if err == nil && prefix.Contains(addr) {
			return true
		}
	}
	return false
}
