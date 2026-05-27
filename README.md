# IP Check

`ipcheck` is a DNS-to-IP resolver CLI focused on unusual domain inputs and fully custom DNS endpoints.

It supports:

- Deep domain names with many dots, such as `a.b.c.d.e.f.example.com`
- IDN domains, such as `例子.测试`
- Explicit root-style FQDN input, such as `example.com.`
- Custom DNS ports for every supported transport
- UDP, TCP, DNS over TLS, DNS over HTTPS, and DNS over QUIC
- Batch input, concurrent queries, JSON/CSV/table output

## Build

This workspace includes a local Go toolchain under `.tools/go`.

```bash
PATH="$PWD/.tools/go/bin:$PATH" go build -o bin/ipcheck ./cmd/ipcheck
```

Run tests:

```bash
PATH="$PWD/.tools/go/bin:$PATH" go test ./...
```

## Basic Usage

```bash
./bin/ipcheck example.com
./bin/ipcheck --type A example.com
./bin/ipcheck --type A,AAAA example.com
```

## GUI

Start the local GUI:

```bash
./bin/ipcheck serve --addr 127.0.0.1:8765
```

Open:

```text
http://127.0.0.1:8765
```

The GUI uses the same resolver core as the CLI and supports the same endpoint formats, custom ports, query types, EDNS0, DNSSEC, DoH method selection, strict domain validation, TLS verification options, batch input, result details, and JSON/CSV export.

## Custom DNS Endpoints

Every protocol supports explicit ports:

```bash
./bin/ipcheck --dns udp://8.8.8.8:5353 example.com
./bin/ipcheck --dns tcp://1.1.1.1:9953 example.com
./bin/ipcheck --dns tls://1.1.1.1:853 example.com
./bin/ipcheck --dns https://dns.google:443/dns-query example.com
./bin/ipcheck --dns quic://dns.adguard-dns.com:853 example.com
```

IPv6 DNS endpoints are supported:

```bash
./bin/ipcheck --dns udp://[2606:4700:4700::1111]:5353 example.com
```

If no port is provided, protocol defaults are used:

| Protocol | Default Port |
| --- | --- |
| UDP | 53 |
| TCP | 53 |
| DoT | 853 |
| DoH | 443 |
| DoQ | 853 |

Supported endpoint schemes:

| Scheme | Meaning |
| --- | --- |
| `udp://` or `dns://` | DNS over UDP |
| `tcp://` | DNS over TCP |
| `tls://` or `dot://` | DNS over TLS |
| `https://` or `doh://` | DNS over HTTPS |
| `http://` | DNS over plain HTTP |
| `quic://` or `doq://` | DNS over QUIC |

## Examples

Compare multiple DNS transports:

```bash
./bin/ipcheck \
  --dns udp://1.1.1.1:53 \
  --dns tcp://1.1.1.1:53 \
  --dns tls://1.1.1.1:853 \
  --dns https://cloudflare-dns.com:443/dns-query \
  --type A,AAAA \
  example.com
```

Use DoH GET:

```bash
./bin/ipcheck --dns https://dns.google:443/dns-query --doh-method GET example.com
```

Custom SNI or certificate debugging:

```bash
./bin/ipcheck --dns 'tls://1.1.1.1:853?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'https://1.1.1.1:443/dns-query?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'tls://127.0.0.1:1853?insecure=1' example.com
```

Batch mode:

```bash
./bin/ipcheck --input domains.txt --dns udp://8.8.8.8:53 --format csv
```

JSON output:

```bash
./bin/ipcheck --dns https://dns.google/dns-query --format json example.com
```

## Domain Handling

The resolver normalizes input before querying:

- URLs are reduced to their host, for example `https://example.com/path` becomes `example.com`
- Ports are stripped from domain input, for example `example.com:443` becomes `example.com`
- IDN names are converted to Punycode
- A trailing root dot is accepted
- Many-dot domains are accepted if each label and the full DNS name fit DNS size limits

Invalid examples:

```text
a..b.com
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com
```

`a..b.com` is rejected because DNS names cannot contain an empty label between two dots. A 64-byte label is rejected because the DNS limit is 63 bytes per label.

By default, underscore labels such as `_sip._tcp.example.com` are allowed with a warning. Use `--strict` to enforce hostname-style labels.

## Useful Options

```text
--dns ENDPOINT             DNS endpoint; repeatable or comma-separated
--type TYPE                DNS query type; repeatable or comma-separated
--input FILE               Read domain names from a file; use - for stdin
--format table|json|csv    Output format
--timeout 3s               Per-request timeout
--retries 1                Retry count for transport errors
--concurrency 16           Concurrent query count
--strict                   Strict hostname validation
--no-edns                  Disable EDNS0
--dnssec                   Set the EDNS DNSSEC DO bit
--doh-method POST|GET      DoH method
--insecure-skip-verify     Skip TLS verification for debugging
```

## Notes

DoQ uses QUIC over UDP, so it can be blocked by firewalls or networks that allow TCP/TLS/HTTPS but block UDP port 853. In that case the tool will report a transport timeout while UDP/TCP/DoH/DoT may still work.
