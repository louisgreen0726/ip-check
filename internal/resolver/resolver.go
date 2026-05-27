package resolver

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"

	"ipcheck/internal/endpoint"
)

func init() {
	if os.Getenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING") == "" {
		_ = os.Setenv("QUIC_GO_DISABLE_RECEIVE_BUFFER_WARNING", "true")
	}
}

type Options struct {
	Timeout            time.Duration
	Retries            int
	EDNS               bool
	DNSSEC             bool
	UDPSize            uint16
	DoHMethod          string
	InsecureSkipVerify bool
}

type ExchangeResult struct {
	Message     *dns.Msg
	Duration    time.Duration
	Protocol    string
	Truncated   bool
	TCPFallback bool
}

type Record struct {
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  uint32 `json:"ttl"`
	Data string `json:"data"`
}

type ParsedMessage struct {
	RCode              string   `json:"rcode"`
	Authoritative      bool     `json:"authoritative"`
	RecursionDesired   bool     `json:"recursionDesired"`
	RecursionAvailable bool     `json:"recursionAvailable"`
	AuthenticatedData  bool     `json:"authenticatedData"`
	CheckingDisabled   bool     `json:"checkingDisabled"`
	Answer             []Record `json:"answer"`
	Authority          []Record `json:"authority"`
	Additional         []Record `json:"additional"`
	CNAMEChain         []string `json:"cnameChain,omitempty"`
	IPs                []string `json:"ips,omitempty"`
}

func DefaultOptions() Options {
	return Options{
		Timeout:   3 * time.Second,
		Retries:   1,
		EDNS:      true,
		DNSSEC:    false,
		UDPSize:   1232,
		DoHMethod: "POST",
	}
}

func QueryType(name string) (uint16, error) {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if upper == "" {
		return 0, fmt.Errorf("查询类型为空")
	}
	if t, ok := dns.StringToType[upper]; ok {
		return t, nil
	}
	return 0, fmt.Errorf("不支持的查询类型: %s", name)
}

func QueryTypeName(qtype uint16) string {
	if name, ok := dns.TypeToString[qtype]; ok {
		return name
	}
	return fmt.Sprintf("TYPE%d", qtype)
}

func NewQuery(fqdn string, qtype uint16, opts Options) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(fqdn), qtype)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	if opts.EDNS {
		size := opts.UDPSize
		if size == 0 {
			size = 1232
		}
		msg.SetEdns0(size, opts.DNSSEC)
	}
	return msg
}

func Exchange(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options) (ExchangeResult, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 3 * time.Second
	}
	if opts.UDPSize == 0 {
		opts.UDPSize = 1232
	}

	var lastErr error
	var lastResult ExchangeResult
	for attempt := 0; attempt <= opts.Retries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
		result, err := exchangeOnce(attemptCtx, ep, query.Copy(), opts)
		cancel()
		if err == nil {
			return result, nil
		}
		lastResult = result
		lastErr = err
	}
	if lastResult.Protocol == "" {
		lastResult.Protocol = string(ep.Scheme)
	}
	return lastResult, lastErr
}

func ParseMessage(msg *dns.Msg) ParsedMessage {
	parsed := ParsedMessage{
		RCode:              dns.RcodeToString[msg.Rcode],
		Authoritative:      msg.Authoritative,
		RecursionDesired:   msg.RecursionDesired,
		RecursionAvailable: msg.RecursionAvailable,
		AuthenticatedData:  msg.AuthenticatedData,
		CheckingDisabled:   msg.CheckingDisabled,
		Answer:             recordsFromRRs(msg.Answer),
		Authority:          recordsFromRRs(msg.Ns),
		Additional:         recordsFromRRs(msg.Extra),
	}

	seenCNAME := map[string]struct{}{}
	for _, rr := range msg.Answer {
		switch v := rr.(type) {
		case *dns.CNAME:
			target := strings.TrimSuffix(v.Target, ".")
			if _, ok := seenCNAME[target]; !ok {
				parsed.CNAMEChain = append(parsed.CNAMEChain, target)
				seenCNAME[target] = struct{}{}
			}
		case *dns.A:
			parsed.IPs = append(parsed.IPs, v.A.String())
		case *dns.AAAA:
			parsed.IPs = append(parsed.IPs, v.AAAA.String())
		}
	}
	return parsed
}

func exchangeOnce(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options) (ExchangeResult, error) {
	switch ep.Scheme {
	case endpoint.UDP:
		return exchangeUDP(ctx, ep, query, opts)
	case endpoint.TCP:
		return exchangeMiekg(ctx, ep, query, opts, "tcp")
	case endpoint.TLS:
		return exchangeMiekg(ctx, ep, query, opts, "tcp-tls")
	case endpoint.HTTPS, endpoint.HTTP:
		return exchangeHTTPS(ctx, ep, query, opts)
	case endpoint.QUIC:
		return exchangeQUIC(ctx, ep, query, opts)
	default:
		return ExchangeResult{}, fmt.Errorf("不支持的 DNS 协议: %s", ep.Scheme)
	}
}

func exchangeUDP(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options) (ExchangeResult, error) {
	start := time.Now()
	client := &dns.Client{
		Net:     "udp",
		Timeout: opts.Timeout,
		UDPSize: opts.UDPSize,
	}
	msg, _, err := client.ExchangeContext(ctx, query, ep.Address())
	if err != nil {
		return ExchangeResult{Protocol: "udp", Duration: time.Since(start)}, err
	}
	if msg.Truncated {
		tcpResult, tcpErr := exchangeMiekg(ctx, ep, query.Copy(), opts, "tcp")
		if tcpErr == nil {
			tcpResult.Duration = time.Since(start)
			tcpResult.Truncated = true
			tcpResult.TCPFallback = true
			tcpResult.Protocol = "udp+tcp"
			return tcpResult, nil
		}
		return ExchangeResult{
			Message:     msg,
			Duration:    time.Since(start),
			Protocol:    "udp",
			Truncated:   true,
			TCPFallback: false,
		}, tcpErr
	}
	return ExchangeResult{
		Message:   msg,
		Duration:  time.Since(start),
		Protocol:  "udp",
		Truncated: msg.Truncated,
	}, nil
}

func exchangeMiekg(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options, netType string) (ExchangeResult, error) {
	start := time.Now()
	client := &dns.Client{
		Net:     netType,
		Timeout: opts.Timeout,
		UDPSize: opts.UDPSize,
	}
	if netType == "tcp-tls" {
		client.TLSConfig = &tls.Config{
			ServerName:         ep.TLSName(),
			InsecureSkipVerify: ep.Insecure || opts.InsecureSkipVerify,
		}
	}
	msg, _, err := client.ExchangeContext(ctx, query, ep.Address())
	protocol := "tcp"
	if netType == "tcp-tls" {
		protocol = "tls"
	}
	if err != nil {
		return ExchangeResult{Protocol: protocol, Duration: time.Since(start)}, err
	}
	return ExchangeResult{
		Message:   msg,
		Duration:  time.Since(start),
		Protocol:  protocol,
		Truncated: msg.Truncated,
	}, nil
}

func exchangeHTTPS(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options) (ExchangeResult, error) {
	start := time.Now()
	wire, err := query.Pack()
	if err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme)}, err
	}

	method := strings.ToUpper(strings.TrimSpace(opts.DoHMethod))
	if method == "" {
		method = "POST"
	}
	if method != "POST" && method != "GET" {
		return ExchangeResult{Protocol: string(ep.Scheme)}, fmt.Errorf("DoH method 只能是 GET 或 POST: %s", opts.DoHMethod)
	}

	endpointURL, err := url.Parse(ep.URL())
	if err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme)}, err
	}

	var body io.Reader
	if method == "GET" {
		q := endpointURL.Query()
		q.Set("dns", base64.RawURLEncoding.EncodeToString(wire))
		endpointURL.RawQuery = q.Encode()
	} else {
		body = bytes.NewReader(wire)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL.String(), body)
	if err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme)}, err
	}
	req.Header.Set("Accept", "application/dns-message")
	if method == "POST" {
		req.Header.Set("Content-Type", "application/dns-message")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName:         ep.TLSName(),
			InsecureSkipVerify: ep.Insecure || opts.InsecureSkipVerify,
		},
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme), Duration: time.Since(start)}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ExchangeResult{Protocol: string(ep.Scheme), Duration: time.Since(start)}, fmt.Errorf("DoH HTTP 状态码异常: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 65536+512))
	if err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme), Duration: time.Since(start)}, err
	}
	msg := new(dns.Msg)
	if err := msg.Unpack(data); err != nil {
		return ExchangeResult{Protocol: string(ep.Scheme), Duration: time.Since(start)}, fmt.Errorf("DoH 响应不是合法 DNS message: %w", err)
	}

	return ExchangeResult{
		Message:   msg,
		Duration:  time.Since(start),
		Protocol:  string(ep.Scheme),
		Truncated: msg.Truncated,
	}, nil
}

func exchangeQUIC(ctx context.Context, ep endpoint.Endpoint, query *dns.Msg, opts Options) (ExchangeResult, error) {
	start := time.Now()
	query.Id = 0
	wire, err := query.Pack()
	if err != nil {
		return ExchangeResult{Protocol: "quic"}, err
	}
	if len(wire) > 65535 {
		return ExchangeResult{Protocol: "quic"}, fmt.Errorf("DNS message 超过 DoQ 2 字节长度限制: %d", len(wire))
	}
	frame := make([]byte, 2+len(wire))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(wire)))
	copy(frame[2:], wire)

	tlsConf := &tls.Config{
		ServerName:         ep.TLSName(),
		NextProtos:         []string{"doq"},
		InsecureSkipVerify: ep.Insecure || opts.InsecureSkipVerify,
	}
	quicConf := &quic.Config{
		HandshakeIdleTimeout: opts.Timeout,
		MaxIdleTimeout:       opts.Timeout,
	}

	conn, err := quic.DialAddr(ctx, ep.Address(), tlsConf, quicConf)
	if err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}
	_ = stream.SetDeadline(time.Now().Add(opts.Timeout))

	if _, err := stream.Write(frame); err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}
	if err := stream.Close(); err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}

	lengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, lengthBuf); err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}
	length := int(binary.BigEndian.Uint16(lengthBuf))
	if length == 0 {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, fmt.Errorf("DoQ 响应长度为 0")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, err
	}

	msg := new(dns.Msg)
	if err := msg.Unpack(data); err != nil {
		return ExchangeResult{Protocol: "quic", Duration: time.Since(start)}, fmt.Errorf("DoQ 响应不是合法 DNS message: %w", err)
	}

	return ExchangeResult{
		Message:   msg,
		Duration:  time.Since(start),
		Protocol:  "quic",
		Truncated: msg.Truncated,
	}, nil
}

func recordsFromRRs(rrs []dns.RR) []Record {
	records := make([]Record, 0, len(rrs))
	for _, rr := range rrs {
		records = append(records, Record{
			Name: strings.TrimSuffix(rr.Header().Name, "."),
			Type: QueryTypeName(rr.Header().Rrtype),
			TTL:  rr.Header().Ttl,
			Data: rrData(rr),
		})
	}
	return records
}

func rrData(rr dns.RR) string {
	switch v := rr.(type) {
	case *dns.A:
		return v.A.String()
	case *dns.AAAA:
		return v.AAAA.String()
	case *dns.CNAME:
		return strings.TrimSuffix(v.Target, ".")
	case *dns.NS:
		return strings.TrimSuffix(v.Ns, ".")
	case *dns.PTR:
		return strings.TrimSuffix(v.Ptr, ".")
	case *dns.MX:
		return fmt.Sprintf("%d %s", v.Preference, strings.TrimSuffix(v.Mx, "."))
	case *dns.SOA:
		return fmt.Sprintf("%s %s %d %d %d %d %d", strings.TrimSuffix(v.Ns, "."), strings.TrimSuffix(v.Mbox, "."), v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.TXT:
		return strings.Join(v.Txt, " ")
	case *dns.SRV:
		return fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, strings.TrimSuffix(v.Target, "."))
	case *dns.CAA:
		return fmt.Sprintf("%d %s %q", v.Flag, v.Tag, v.Value)
	default:
		text := rr.String()
		fields := strings.Fields(text)
		if len(fields) > 4 {
			return strings.Join(fields[4:], " ")
		}
		return text
	}
}

func IsNoAnswer(parsed ParsedMessage) bool {
	return parsed.RCode == "NOERROR" && len(parsed.Answer) == 0
}

func LookupHostPort(host string, port int) string {
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}
