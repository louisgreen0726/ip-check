package endpoint

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Scheme string

const (
	UDP   Scheme = "udp"
	TCP   Scheme = "tcp"
	TLS   Scheme = "tls"
	HTTPS Scheme = "https"
	HTTP  Scheme = "http"
	QUIC  Scheme = "quic"
)

type Endpoint struct {
	Raw        string            `json:"raw"`
	Scheme     Scheme            `json:"scheme"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Path       string            `json:"path,omitempty"`
	ServerName string            `json:"serverName,omitempty"`
	Insecure   bool              `json:"insecure,omitempty"`
	Query      map[string]string `json:"query,omitempty"`
}

func Parse(raw string) (Endpoint, error) {
	original := strings.TrimSpace(raw)
	if original == "" {
		return Endpoint{}, fmt.Errorf("DNS endpoint 为空")
	}

	if !strings.Contains(original, "://") {
		return parseHostPort(original, UDP, original)
	}

	u, err := url.Parse(original)
	if err != nil {
		return Endpoint{}, fmt.Errorf("DNS endpoint URL 解析失败: %w", err)
	}

	scheme, err := normalizeScheme(u.Scheme)
	if err != nil {
		return Endpoint{}, err
	}

	host := u.Hostname()
	if host == "" {
		return Endpoint{}, fmt.Errorf("DNS endpoint 缺少 host: %s", original)
	}

	port := defaultPort(scheme)
	if u.Port() != "" {
		port, err = parsePort(u.Port())
		if err != nil {
			return Endpoint{}, err
		}
	}

	path := u.EscapedPath()
	if (scheme == HTTPS || scheme == HTTP) && path == "" {
		path = "/dns-query"
	}

	values := u.Query()
	serverName := values.Get("sni")
	if serverName == "" {
		serverName = values.Get("server_name")
	}
	insecure := parseBool(values.Get("insecure")) || parseBool(values.Get("skip_verify"))
	values.Del("sni")
	values.Del("server_name")
	values.Del("insecure")
	values.Del("skip_verify")

	query := map[string]string{}
	for k, vals := range values {
		if len(vals) > 0 {
			query[k] = vals[len(vals)-1]
		}
	}

	return Endpoint{
		Raw:        original,
		Scheme:     scheme,
		Host:       host,
		Port:       port,
		Path:       path,
		ServerName: serverName,
		Insecure:   insecure,
		Query:      query,
	}, nil
}

func MustParseMany(raws []string) ([]Endpoint, error) {
	endpoints := make([]Endpoint, 0, len(raws))
	for _, raw := range raws {
		ep, err := Parse(raw)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, nil
}

func (e Endpoint) Address() string {
	return net.JoinHostPort(e.Host, strconv.Itoa(e.Port))
}

func (e Endpoint) URL() string {
	if e.Scheme != HTTPS && e.Scheme != HTTP {
		return ""
	}
	u := url.URL{
		Scheme: string(e.Scheme),
		Host:   e.Address(),
		Path:   e.Path,
	}
	q := u.Query()
	for k, v := range e.Query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (e Endpoint) Display() string {
	if e.Raw != "" {
		return e.Raw
	}
	return fmt.Sprintf("%s://%s", e.Scheme, e.Address())
}

func (e Endpoint) TLSName() string {
	if e.ServerName != "" {
		return e.ServerName
	}
	return e.Host
}

func parseHostPort(raw string, scheme Scheme, original string) (Endpoint, error) {
	host := raw
	port := defaultPort(scheme)

	if h, p, err := net.SplitHostPort(raw); err == nil {
		parsedPort, err := parsePort(p)
		if err != nil {
			return Endpoint{}, err
		}
		host = h
		port = parsedPort
	} else if idx := strings.LastIndex(raw, ":"); idx > -1 && strings.Count(raw, ":") == 1 {
		if parsedPort, err := parsePort(raw[idx+1:]); err == nil {
			host = raw[:idx]
			port = parsedPort
		}
	}

	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	if host == "" {
		return Endpoint{}, fmt.Errorf("DNS endpoint 缺少 host: %s", original)
	}

	return Endpoint{
		Raw:    original,
		Scheme: scheme,
		Host:   host,
		Port:   port,
	}, nil
}

func normalizeScheme(s string) (Scheme, error) {
	switch strings.ToLower(s) {
	case "dns", "udp":
		return UDP, nil
	case "tcp":
		return TCP, nil
	case "tls", "dot":
		return TLS, nil
	case "https", "doh":
		return HTTPS, nil
	case "http":
		return HTTP, nil
	case "quic", "doq":
		return QUIC, nil
	default:
		return "", fmt.Errorf("不支持的 DNS 协议: %s", s)
	}
}

func defaultPort(s Scheme) int {
	switch s {
	case UDP, TCP:
		return 53
	case TLS, QUIC:
		return 853
	case HTTPS:
		return 443
	case HTTP:
		return 80
	default:
		return 53
	}
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("端口必须是整数: %q", raw)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("端口超出范围 1-65535: %d", port)
	}
	return port, nil
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
