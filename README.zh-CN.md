# IP Check

[English](README.md)

`ipcheck` 是一个用 Go 编写的 DNS 解析检查工具，包含命令行和本地 Web GUI。它适合用来验证域名在不同 DNS 端点、不同传输协议下的解析结果，也能处理 IDN、显式根点 FQDN、URL、IPv6 DNS 端点和多级子域名等输入。

当前项目只发布 Go 命令本体。Web GUI 由同一个二进制文件在本地提供服务，不需要 Electron 套壳、Android 工程或移动端构建步骤。

## 功能

- 支持 `A`、`AAAA`、`CNAME`、`MX`、`NS`、`TXT`、`SOA`、`CAA`、`SRV`、`PTR`、`HTTPS`、`SVCB` 记录
- 支持 UDP、TCP、DNS over TLS、DNS over HTTPS、普通 HTTP DoH 风格端点和 DNS over QUIC
- 可配置自定义 DNS 主机、端口、SNI 和 TLS 校验行为
- 自动规范化 URL、`host:port`、IDN 域名和带根点的 FQDN
- 支持批量查询、超时、重试和并发控制
- CLI 支持 table、JSON、CSV 输出
- 本地 Web GUI 支持筛选、详情查看、JSON/CSV 导出，以及可选的 IP 地理位置、ASN、运营商信息补充

## 构建与测试

仓库内包含本地 Go 工具链 `.tools/go`。如果希望使用项目预期的工具链，可以这样运行：

```bash
PATH="$PWD/.tools/go/bin:$PATH" go test ./...
PATH="$PWD/.tools/go/bin:$PATH" go build -o bin/ipcheck ./cmd/ipcheck
```

如果系统 Go 版本与 `go.mod` 匹配，也可以直接使用普通 `go` 命令：

```bash
go test ./...
go build -o bin/ipcheck ./cmd/ipcheck
```

## 命令行使用

```bash
./bin/ipcheck example.com
./bin/ipcheck --type A example.com
./bin/ipcheck --type A,AAAA --dns udp://1.1.1.1:53 example.com
```

批量输入：

```bash
./bin/ipcheck --input domains.txt --dns udp://8.8.8.8:53 --format csv
```

JSON 输出并补充 IP 信息：

```bash
./bin/ipcheck --dns https://dns.google/dns-query --format json --ip-info example.com
```

## 本地 GUI

启动 GUI 服务：

```bash
./bin/ipcheck serve --addr 127.0.0.1:8765
```

打开：

```text
http://127.0.0.1:8765
```

也可以让命令自动打开浏览器：

```bash
./bin/ipcheck serve --addr 127.0.0.1:8765 --open
```

GUI 使用与 CLI 相同的解析核心，支持相同的 DNS 端点格式、查询类型、EDNS0、DNSSEC、DoH 方法选择、严格校验、TLS 调试选项、批量输入、结果详情、筛选和 JSON/CSV 导出。

## DNS 端点

每种协议都支持显式端口：

```bash
./bin/ipcheck --dns udp://8.8.8.8:5353 example.com
./bin/ipcheck --dns tcp://1.1.1.1:9953 example.com
./bin/ipcheck --dns tls://1.1.1.1:853 example.com
./bin/ipcheck --dns https://dns.google:443/dns-query example.com
./bin/ipcheck --dns quic://dns.adguard-dns.com:853 example.com
```

支持 IPv6 DNS 端点：

```bash
./bin/ipcheck --dns udp://[2606:4700:4700::1111]:5353 example.com
```

支持的端点 scheme：

| Scheme | 含义 |
| --- | --- |
| `udp://` 或 `dns://` | DNS over UDP |
| `tcp://` | DNS over TCP |
| `tls://` 或 `dot://` | DNS over TLS |
| `https://` 或 `doh://` | DNS over HTTPS |
| `http://` | 普通 HTTP |
| `quic://` 或 `doq://` | DNS over QUIC |

默认端口：

| 协议 | 默认端口 |
| --- | --- |
| UDP | 53 |
| TCP | 53 |
| DoT | 853 |
| DoH | 443 |
| DoQ | 853 |

DoH GET、自定义 SNI 和跳过 TLS 校验示例：

```bash
./bin/ipcheck --dns https://dns.google:443/dns-query --doh-method GET example.com
./bin/ipcheck --dns 'tls://1.1.1.1:853?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'https://1.1.1.1:443/dns-query?sni=cloudflare-dns.com' example.com
./bin/ipcheck --dns 'tls://127.0.0.1:1853?insecure=1' example.com
```

## 域名处理

查询前，`ipcheck` 会先规范化域名输入：

- URL 会被还原为主机名，例如 `https://example.com/path` 会变成 `example.com`
- 域名输入中的端口会被移除，例如 `example.com:443` 会变成 `example.com`
- IDN 域名会转换为 Punycode
- 接受末尾带根点的 FQDN
- 只要单个 label 和完整 DNS 名称符合 DNS 长度限制，就接受多级子域名

无效示例：

```text
a..b.com
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com
```

`a..b.com` 会被拒绝，因为两个点之间不能出现空 label。64 字节 label 会被拒绝，因为 DNS 单个 label 的长度限制是 63 字节。

默认情况下，`_sip._tcp.example.com` 这类带下划线的 label 会被允许并给出警告。使用 `--strict` 可以启用更严格的主机名校验。

## 常用参数

```text
--dns ENDPOINT             DNS 端点；可重复或用逗号分隔
--type TYPE                DNS 查询类型；可重复或用逗号分隔
--input FILE               从文件读取域名；使用 - 表示 stdin
--format table|json|csv    输出格式
--timeout 3s               单次请求超时
--retries 1                传输错误重试次数
--concurrency 16           并发查询数
--strict                   严格主机名校验
--no-edns                  禁用 EDNS0
--dnssec                   设置 EDNS DNSSEC DO bit
--doh-method POST|GET      DoH 方法
--insecure-skip-verify     跳过 TLS 证书校验，仅建议调试使用
--ip-info                  补充解析 IP 的地理位置、ASN 和运营商信息
--version                  输出版本
--examples                 输出 CLI 示例
```

## 项目结构

```text
cmd/ipcheck/          CLI 入口和本地 GUI 服务
cmd/ipcheck/web/      嵌入式 Web GUI 资源
internal/domain/      域名规范化与校验
internal/endpoint/    DNS 端点解析
internal/resolver/    DNS 查询与响应解析
internal/ipinfo/      可选的 IP 元数据查询
```

## CI

GitHub Actions 会运行测试、vet 和 Linux CLI 构建。当前 workflow 只上传普通的 `ipcheck` 二进制文件；桌面端和移动端打包不属于当前项目范围。

## 说明

DNS over QUIC 使用 UDP，因此可能会被只允许 TCP、TLS 或 HTTPS 的网络阻断。如果 UDP 853 端口不可用，`ipcheck` 会报告传输超时，而 UDP、TCP、DoH 或 DoT 的其他检查可能仍然正常。
