package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
	"unicode"

	"ipcheck/internal/domain"
	"ipcheck/internal/endpoint"
	"ipcheck/internal/ipinfo"
	"ipcheck/internal/resolver"
)

const version = "0.1.0"

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	for _, part := range splitComma(value) {
		if part != "" {
			*s = append(*s, part)
		}
	}
	return nil
}

type cliOptions struct {
	dns                stringList
	qtypes             stringList
	inputFile          string
	format             string
	timeout            time.Duration
	retries            int
	concurrency        int
	strict             bool
	noEDNS             bool
	dnssec             bool
	dohMethod          string
	insecureSkipVerify bool
	ipInfo             bool
	version            bool
	examples           bool
}

type result struct {
	Input             string                  `json:"input"`
	Domain            string                  `json:"domain,omitempty"`
	ASCII             string                  `json:"ascii,omitempty"`
	FQDN              string                  `json:"fqdn,omitempty"`
	Type              string                  `json:"type,omitempty"`
	Resolver          string                  `json:"resolver,omitempty"`
	ResolverProtocol  string                  `json:"resolverProtocol,omitempty"`
	TransportProtocol string                  `json:"transportProtocol,omitempty"`
	Status            string                  `json:"status"`
	RCode             string                  `json:"rcode,omitempty"`
	DurationMS        int64                   `json:"durationMs,omitempty"`
	Truncated         bool                    `json:"truncated,omitempty"`
	TCPFallback       bool                    `json:"tcpFallback,omitempty"`
	Warnings          []string                `json:"warnings,omitempty"`
	Error             string                  `json:"error,omitempty"`
	IPs               []string                `json:"ips,omitempty"`
	IPInfo            []ipinfo.Info           `json:"ipInfo,omitempty"`
	CNAMEChain        []string                `json:"cnameChain,omitempty"`
	Answer            []resolver.Record       `json:"answer,omitempty"`
	Authority         []resolver.Record       `json:"authority,omitempty"`
	Additional        []resolver.Record       `json:"additional,omitempty"`
	Response          *resolver.ParsedMessage `json:"response,omitempty"`
}

type task struct {
	name  domain.Name
	ep    endpoint.Endpoint
	qtype uint16
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, stdin *os.File) error {
	if len(args) > 0 && args[0] == "serve" {
		return runServer(args[1:], stdout, stderr)
	}

	opts, remaining := parseFlags(args, stderr)
	if opts.version {
		fmt.Fprintf(stdout, "ipcheck %s\n", version)
		return nil
	}
	if opts.examples {
		printExamples(stdout)
		return nil
	}

	domains, err := collectInputs(remaining, opts.inputFile, stdin)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		return errors.New("请提供至少一个域名，或使用 --input 从文件读取")
	}

	if len(opts.dns) == 0 {
		opts.dns = stringList{"udp://1.1.1.1:53"}
	}
	if len(opts.qtypes) == 0 {
		opts.qtypes = stringList{"A", "AAAA"}
	}

	endpoints, err := endpoint.MustParseMany(opts.dns)
	if err != nil {
		return err
	}

	qtypes := make([]uint16, 0, len(opts.qtypes))
	for _, qtypeRaw := range opts.qtypes {
		qtype, err := resolver.QueryType(qtypeRaw)
		if err != nil {
			return err
		}
		qtypes = append(qtypes, qtype)
	}

	results := resolveAll(context.Background(), domains, endpoints, qtypes, opts)
	if opts.ipInfo {
		enrichIPInfo(context.Background(), results)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Input != results[j].Input {
			return results[i].Input < results[j].Input
		}
		if results[i].Resolver != results[j].Resolver {
			return results[i].Resolver < results[j].Resolver
		}
		return results[i].Type < results[j].Type
	})

	switch strings.ToLower(opts.format) {
	case "json":
		return writeJSON(stdout, results)
	case "csv":
		return writeCSV(stdout, results)
	case "table", "":
		return writeTable(stdout, results)
	default:
		return fmt.Errorf("不支持的输出格式: %s", opts.format)
	}
}

func parseFlags(args []string, stderr io.Writer) (cliOptions, []string) {
	opts := cliOptions{}
	fs := flag.NewFlagSet("ipcheck", flag.ExitOnError)
	fs.SetOutput(stderr)
	fs.Var(&opts.dns, "dns", "DNS endpoint，可重复或逗号分隔，例如 udp://8.8.8.8:5353,tls://1.1.1.1:853,https://dns.google:443/dns-query,quic://dns.adguard-dns.com:853")
	fs.Var(&opts.qtypes, "type", "查询类型，可重复或逗号分隔，默认 A,AAAA")
	fs.StringVar(&opts.inputFile, "input", "", "从文件读取域名；使用 - 表示 stdin")
	fs.StringVar(&opts.format, "format", "table", "输出格式: table, json, csv")
	fs.DurationVar(&opts.timeout, "timeout", 3*time.Second, "单次请求超时")
	fs.IntVar(&opts.retries, "retries", 1, "网络错误重试次数")
	fs.IntVar(&opts.concurrency, "concurrency", 16, "并发查询数")
	fs.BoolVar(&opts.strict, "strict", false, "严格主机名校验模式")
	fs.BoolVar(&opts.noEDNS, "no-edns", false, "禁用 EDNS0")
	fs.BoolVar(&opts.dnssec, "dnssec", false, "设置 EDNS DNSSEC DO bit")
	fs.StringVar(&opts.dohMethod, "doh-method", "POST", "DoH 方法: POST 或 GET")
	fs.BoolVar(&opts.insecureSkipVerify, "insecure-skip-verify", false, "跳过 TLS 证书校验，仅建议调试使用")
	fs.BoolVar(&opts.ipInfo, "ip-info", false, "查询解析出的 IP 位置、ASN 和运营商信息")
	fs.BoolVar(&opts.version, "version", false, "显示版本")
	fs.BoolVar(&opts.examples, "examples", false, "显示示例")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: ipcheck [options] <domain...>\n\n")
		fmt.Fprintln(fs.Output(), "Examples:")
		fmt.Fprintln(fs.Output(), "  ipcheck example.com")
		fmt.Fprintln(fs.Output(), "  ipcheck --dns udp://8.8.8.8:5353 --type A example.com")
		fmt.Fprintln(fs.Output(), "  ipcheck --dns https://dns.google:443/dns-query --type A,AAAA example.com")
		fmt.Fprintln(fs.Output(), "  ipcheck --dns tls://1.1.1.1:853 --dns quic://dns.adguard-dns.com:853 example.com")
		fmt.Fprintln(fs.Output(), "  ipcheck serve --addr 127.0.0.1:8765")
		fmt.Fprintln(fs.Output(), "\nOptions:")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)
	if opts.concurrency < 1 {
		opts.concurrency = 1
	}
	return opts, fs.Args()
}

func collectInputs(args []string, inputFile string, stdin *os.File) ([]string, error) {
	items := make([]string, 0)
	items = append(items, args...)

	if inputFile != "" {
		var data []byte
		var err error
		if inputFile == "-" {
			data, err = io.ReadAll(stdin)
		} else {
			data, err = os.ReadFile(inputFile)
		}
		if err != nil {
			return nil, err
		}
		items = append(items, splitInputs(string(data))...)
	} else if stdin != nil {
		if stat, err := stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(stdin)
			if err != nil {
				return nil, err
			}
			items = append(items, splitInputs(string(data))...)
		}
	}

	seen := map[string]struct{}{}
	unique := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, item)
	}
	return unique, nil
}

func resolveAll(ctx context.Context, inputs []string, endpoints []endpoint.Endpoint, qtypes []uint16, cli cliOptions) []result {
	all := make([]result, 0, len(inputs)*len(endpoints)*len(qtypes))
	taskList := make([]task, 0, len(inputs)*len(endpoints)*len(qtypes))
	tasks := make(chan task)
	results := make(chan result)

	for _, input := range inputs {
		name, err := domain.Normalize(input, domain.Options{Strict: cli.strict})
		if err != nil {
			all = append(all, result{
				Input:  input,
				Domain: name.Host,
				Status: "INVALID_DOMAIN",
				Error:  err.Error(),
			})
			continue
		}
		for _, ep := range endpoints {
			for _, qtype := range qtypes {
				taskList = append(taskList, task{name: name, ep: ep, qtype: qtype})
			}
		}
	}

	workerCount := cli.concurrency
	if len(taskList) < workerCount {
		workerCount = len(taskList)
	}
	if workerCount == 0 {
		close(tasks)
		close(results)
		return all
	}

	resolverOpts := resolver.DefaultOptions()
	resolverOpts.Timeout = cli.timeout
	resolverOpts.Retries = cli.retries
	resolverOpts.EDNS = !cli.noEDNS
	resolverOpts.DNSSEC = cli.dnssec
	resolverOpts.DoHMethod = cli.dohMethod
	resolverOpts.InsecureSkipVerify = cli.insecureSkipVerify

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for t := range tasks {
				results <- resolveOne(ctx, t, resolverOpts)
			}
		}()
	}

	go func() {
		for _, t := range taskList {
			tasks <- t
		}
		close(tasks)
		wg.Wait()
		close(results)
	}()

	for res := range results {
		all = append(all, res)
	}

	return all
}

func resolveOne(ctx context.Context, t task, opts resolver.Options) result {
	qtypeName := resolver.QueryTypeName(t.qtype)
	base := result{
		Input:            t.name.Original,
		Domain:           t.name.Host,
		ASCII:            t.name.ASCII,
		FQDN:             t.name.FQDN,
		Type:             qtypeName,
		Resolver:         t.ep.Display(),
		ResolverProtocol: string(t.ep.Scheme),
		Warnings:         t.name.Warnings,
	}

	query := resolver.NewQuery(t.name.FQDN, t.qtype, opts)
	exchange, err := resolver.Exchange(ctx, t.ep, query, opts)
	if err != nil {
		base.Status = "ERROR"
		base.TransportProtocol = exchange.Protocol
		base.DurationMS = exchange.Duration.Milliseconds()
		base.Error = err.Error()
		return base
	}

	parsed := resolver.ParseMessage(exchange.Message)
	base.Response = &parsed
	base.TransportProtocol = exchange.Protocol
	base.Status = "OK"
	base.RCode = parsed.RCode
	base.DurationMS = exchange.Duration.Milliseconds()
	base.Truncated = exchange.Truncated
	base.TCPFallback = exchange.TCPFallback
	base.IPs = parsed.IPs
	base.CNAMEChain = parsed.CNAMEChain
	base.Answer = parsed.Answer
	base.Authority = parsed.Authority
	base.Additional = parsed.Additional

	if parsed.RCode != "NOERROR" {
		base.Status = "DNS_ERROR"
	} else if resolver.IsNoAnswer(parsed) {
		base.Status = "NO_ANSWER"
	}

	return base
}

func writeJSON(w io.Writer, results []result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func writeCSV(w io.Writer, results []result) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"input", "domain", "ascii", "type", "resolver", "protocol", "status", "rcode", "ips", "location", "operator", "answers", "duration_ms", "error", "warnings"}); err != nil {
		return err
	}
	for _, r := range results {
		if err := cw.Write([]string{
			r.Input,
			r.Domain,
			r.ASCII,
			r.Type,
			r.Resolver,
			r.TransportProtocol,
			r.Status,
			r.RCode,
			strings.Join(r.IPs, ";"),
			ipLocationsInline(r.IPInfo),
			ipOperatorsInline(r.IPInfo),
			recordsInline(r.Answer),
			fmt.Sprintf("%d", r.DurationMS),
			r.Error,
			strings.Join(r.Warnings, ";"),
		}); err != nil {
			return err
		}
	}
	return cw.Error()
}

func writeTable(w io.Writer, results []result) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "INPUT\tTYPE\tDNS\tPROTO\tSTATUS\tRCODE\tIPS/ANSWER\tLOCATION\tOPERATOR\tMS\tERROR")
	for _, r := range results {
		answer := strings.Join(r.IPs, ",")
		if answer == "" {
			answer = recordsInline(r.Answer)
		}
		if answer == "" && len(r.CNAMEChain) > 0 {
			answer = "CNAME " + strings.Join(r.CNAMEChain, " -> ")
		}
		if len(answer) > 80 {
			answer = answer[:77] + "..."
		}
		errMsg := r.Error
		if len(errMsg) > 80 {
			errMsg = errMsg[:77] + "..."
		}
		location := ipLocationsInline(r.IPInfo)
		operator := ipOperatorsInline(r.IPInfo)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			r.Input,
			r.Type,
			r.Resolver,
			r.TransportProtocol,
			r.Status,
			r.RCode,
			answer,
			location,
			operator,
			r.DurationMS,
			errMsg,
		)
	}
	return tw.Flush()
}

func recordsInline(records []resolver.Record) string {
	if len(records) == 0 {
		return ""
	}
	parts := make([]string, 0, len(records))
	for _, rec := range records {
		parts = append(parts, fmt.Sprintf("%s=%s", rec.Type, rec.Data))
	}
	return strings.Join(parts, ";")
}

func enrichIPInfo(ctx context.Context, results []result) {
	ips := make([]string, 0)
	for _, res := range results {
		ips = append(ips, res.IPs...)
	}
	infos := ipinfo.NewClient().LookupMany(ctx, ips)
	for idx := range results {
		for _, ip := range results[idx].IPs {
			if info, ok := infos[ip]; ok {
				results[idx].IPInfo = append(results[idx].IPInfo, info)
			}
		}
	}
}

func ipLocationsInline(infos []ipinfo.Info) string {
	parts := make([]string, 0, len(infos))
	for _, info := range infos {
		if info.Error != "" {
			continue
		}
		location := strings.Join(nonEmpty(info.Country, info.Region, info.City), "/")
		if location != "" {
			parts = append(parts, info.IP+" "+location)
		}
	}
	return strings.Join(parts, ";")
}

func ipOperatorsInline(infos []ipinfo.Info) string {
	parts := make([]string, 0, len(infos))
	for _, info := range infos {
		if info.Error != "" {
			continue
		}
		operator := strings.Join(nonEmpty(info.ASN, firstNonEmpty(info.ISP, info.Org)), " ")
		if operator != "" {
			parts = append(parts, info.IP+" "+operator)
		}
	}
	return strings.Join(parts, ";")
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func splitComma(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func splitInputs(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
}

func printExamples(w io.Writer) {
	fmt.Fprintln(w, "Common examples:")
	fmt.Fprintln(w, "  ipcheck example.com")
	fmt.Fprintln(w, "  ipcheck --dns udp://8.8.8.8:5353 example.com")
	fmt.Fprintln(w, "  ipcheck --dns tcp://1.1.1.1:9953 --type A example.com")
	fmt.Fprintln(w, "  ipcheck --dns tls://1.1.1.1:853 --type A,AAAA example.com")
	fmt.Fprintln(w, "  ipcheck --dns https://dns.google:443/dns-query --doh-method POST example.com")
	fmt.Fprintln(w, "  ipcheck --dns quic://dns.adguard-dns.com:853 example.com")
	fmt.Fprintln(w, "  ipcheck --dns udp://[2606:4700:4700::1111]:5353 example.com")
	fmt.Fprintln(w, "  ipcheck --input domains.txt --dns udp://1.1.1.1:53 --format csv")
	fmt.Fprintln(w, "  ipcheck serve --addr 127.0.0.1:8765")
}
