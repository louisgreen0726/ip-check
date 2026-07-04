# IP Check

[中文文档](README.zh-CN.md)

`ipcheck` is a Go DNS resolver CLI with a local web GUI. It is built for checking
how domains resolve through custom DNS endpoints, including uncommon inputs such
as IDNs, explicit root-dot FQDNs, URLs, IPv6 endpoints, and deeply nested names.

The project currently ships as a Go command. The browser GUI is served by the
same binary, so no Electron wrapper, Android project, or mobile build step is
required.

## Features

- Resolve `A`, `AAAA`, `CNAME`, `MX`, `NS`, `TXT`, `SOA`, `CAA`, `SRV`, `PTR`,
  `HTTPS`, and `SVCB` records
- Query UDP, TCP, DNS over TLS, DNS over HTTPS, plain HTTP DoH-style endpoints,
  and DNS over QUIC
- Configure custom DNS hosts, ports, SNI, and TLS verification behavior
- Normalize URLs, `host:port` values, IDN domains, and root-dot FQDN input
- Run batch queries with configurable timeout, retries, and concurrency
- Export CLI results as table, JSON, or CSV
- Use a local web GUI with filtering, details, JSON/CSV export, and optional IP
  location/ASN/operator enrichment from `ipapi.co`

## Build And Test

This workspace includes a local Go toolchain under `.tools/go`. Use it when you
want the exact toolchain expected by this repository:

```bash
PATH="$PWD/.tools/go/bin:$PATH" go test ./...
PATH="$PWD/.tools/go/bin:$PATH" go build -o bin/ipcheck ./cmd/ipcheck
```

If your system Go version matches `go.mod`, the regular `go` command works too:

```bash
go test ./...
go build -o bin/ipcheck ./cmd/ipcheck
```

## CLI Usage

```bash
./bin/ipcheck example.com
./bin/ipcheck --type A example.com
./bin/ipcheck --type A,AAAA --dns udp://1.1.1.1:53 example.com
```

Batch input:

```bash
./bin/ipcheck --input domains.txt --dns udp://8.8.8.8:53 --format csv
```

JSON output with IP enrichment:

```bash
./bin/ipcheck --dns https://dns.google/dns-query --format json --ip-info example.com
```

## Local GUI

Start the GUI and open the browser:

```bash
./bin/ipcheck
```

The default local URL is:

```text
http://127.0.0.1:8765
```

Use `serve` when you want to customize the listening address:

```bash
./bin/ipcheck serve --addr 127.0.0.1:8765 --open
```

The GUI uses the same resolver core as the CLI. It supports the same endpoint
formats, query types, EDNS0, DNSSEC, DoH method selection, strict validation,
TLS debugging options, batch input, result details, filtering, and JSON/CSV
export.

## DNS Endpoints

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

Supported endpoint schemes:

| Scheme | Meaning |
| --- | --- |
| `udp://` or `dns://` | DNS over UDP |
| `tcp://` | DNS over TCP |
| `tls://` or `dot://` | DNS over TLS |
| `https://` or `doh://` | DNS over HTTPS |
| `http://` | DNS over plain HTTP |
| `quic://` or `doq://` | DNS over QUIC |

Default ports:

| Protocol | Default Port |
| --- | --- |
| UDP | 53 |
| TCP | 53 |
| DoT | 853 |
| DoH | 443 |
| DoQ | 853 |

Examples with DoH GET, custom SNI, and insecure TLS debugging:

```bash
./bin/ipcheck --dns https://dns.google:443/dns-query --doh-method GET example.com
./bin/ipcheck --dns 'tls://1.1.1.1:853?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'https://1.1.1.1:443/dns-query?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'tls://127.0.0.1:1853?insecure=1' example.com
```

## Domain Handling

Before querying, `ipcheck` normalizes domain input:

- URLs are reduced to their host, such as `https://example.com/path` to
  `example.com`
- Ports are stripped from domain input, such as `example.com:443` to
  `example.com`
- IDN names are converted to Punycode
- A trailing root dot is accepted
- Many-dot domains are accepted when every label and the full DNS name fit DNS
  size limits

Invalid examples:

```text
a..b.com
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com
```

`a..b.com` is rejected because DNS names cannot contain an empty label between
two dots. A 64-byte label is rejected because the DNS label limit is 63 bytes.

By default, underscore labels such as `_sip._tcp.example.com` are allowed with a
warning. Use `--strict` to enforce hostname-style labels.

## Options

```text
--dns ENDPOINT             DNS endpoint; repeatable or comma-separated
--type TYPE                DNS query type; repeatable or comma-separated
--input FILE               Read domain names from a file; use - for stdin
--format table|json|csv    Output format
--timeout 3s               Per-request timeout, from 100ms to 30s
--retries 1                Retry count for transport errors, from 0 to 5
--concurrency 16           Concurrent query count, from 1 to 128
--strict                   Strict hostname validation
--no-edns                  Disable EDNS0
--dnssec                   Set the EDNS DNSSEC DO bit
--doh-method POST|GET      DoH method
--insecure-skip-verify     Skip TLS verification for debugging
--ip-info                  Enrich resolved IPs with location, ASN, and operator data
--version                  Print version
--examples                 Print CLI examples
```

## Project Layout

```text
cmd/ipcheck/          CLI entry point and local GUI server
cmd/ipcheck/web/      Embedded web GUI assets
internal/domain/      Domain normalization and validation
internal/endpoint/    DNS endpoint parsing
internal/resolver/    DNS query and response parsing
internal/ipinfo/      Optional IP metadata lookup
```

## CI

GitHub Actions runs tests and vet, then uploads binary artifacts for
`linux-amd64`, `linux-arm64`, `windows-amd64`, and `windows-arm64`. Desktop and
mobile packaging are intentionally outside the current project scope.

## Notes

DNS over QUIC uses UDP, so it can be blocked by networks that allow TCP, TLS, or
HTTPS but block UDP port 853. In that case `ipcheck` will report a transport
timeout while UDP, TCP, DoH, or DoT checks may still work.
