package mobilecore

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"ipcheck/internal/domain"
	"ipcheck/internal/endpoint"
	"ipcheck/internal/resolver"
)

const Version = "0.1.0"

type Request struct {
	Domains            []string `json:"domains"`
	DomainText         string   `json:"domainText"`
	DNS                []string `json:"dns"`
	Types              []string `json:"types"`
	TimeoutMS          int      `json:"timeoutMs"`
	Retries            int      `json:"retries"`
	Concurrency        int      `json:"concurrency"`
	Strict             bool     `json:"strict"`
	EDNS               *bool    `json:"edns"`
	DNSSEC             bool     `json:"dnssec"`
	DoHMethod          string   `json:"dohMethod"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
}

type Response struct {
	Results []Result `json:"results,omitempty"`
	Summary Summary  `json:"summary,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type Summary struct {
	Total      int `json:"total"`
	OK         int `json:"ok"`
	NoAnswer   int `json:"noAnswer"`
	DNSError   int `json:"dnsError"`
	Error      int `json:"error"`
	Invalid    int `json:"invalid"`
	WithIPs    int `json:"withIps"`
	DurationMS int `json:"durationMs"`
}

type Result struct {
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

// ResolveJSON is the Android bridge entry point. It returns a JSON Response.
func ResolveJSON(requestJSON string) string {
	var req Request
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return marshal(Response{Error: "JSON 请求解析失败: " + err.Error()})
	}

	start := time.Now()
	results, err := Resolve(context.Background(), req)
	if err != nil {
		return marshal(Response{Error: err.Error()})
	}
	return marshal(Response{
		Results: results,
		Summary: summarize(results, time.Since(start)),
	})
}

func HealthJSON() string {
	return marshal(map[string]string{"status": "ok", "version": Version})
}

func Resolve(ctx context.Context, req Request) ([]Result, error) {
	inputs := uniqueStrings(append(req.Domains, splitInputs(req.DomainText)...))
	if len(inputs) == 0 {
		return nil, errors.New("请至少输入一个域名")
	}

	dnsList := uniqueStrings(req.DNS)
	if len(dnsList) == 0 {
		dnsList = []string{"udp://1.1.1.1:53"}
	}
	endpoints, err := endpoint.MustParseMany(dnsList)
	if err != nil {
		return nil, err
	}

	typeNames := uniqueStrings(req.Types)
	if len(typeNames) == 0 {
		typeNames = []string{"A", "AAAA"}
	}
	qtypes := make([]uint16, 0, len(typeNames))
	for _, typeName := range typeNames {
		qtype, err := resolver.QueryType(typeName)
		if err != nil {
			return nil, err
		}
		qtypes = append(qtypes, qtype)
	}

	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	concurrency := req.Concurrency
	if concurrency < 1 {
		concurrency = 16
	}
	edns := true
	if req.EDNS != nil {
		edns = *req.EDNS
	}
	opts := resolver.DefaultOptions()
	opts.Timeout = timeout
	opts.Retries = req.Retries
	opts.EDNS = edns
	opts.DNSSEC = req.DNSSEC
	opts.DoHMethod = strings.ToUpper(strings.TrimSpace(req.DoHMethod))
	if opts.DoHMethod == "" {
		opts.DoHMethod = "POST"
	}
	opts.InsecureSkipVerify = req.InsecureSkipVerify

	results := resolveAll(ctx, inputs, endpoints, qtypes, req.Strict, concurrency, opts)
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Input != results[j].Input {
			return results[i].Input < results[j].Input
		}
		if results[i].Resolver != results[j].Resolver {
			return results[i].Resolver < results[j].Resolver
		}
		return results[i].Type < results[j].Type
	})
	return results, nil
}

func resolveAll(ctx context.Context, inputs []string, endpoints []endpoint.Endpoint, qtypes []uint16, strict bool, concurrency int, opts resolver.Options) []Result {
	all := make([]Result, 0, len(inputs)*len(endpoints)*len(qtypes))
	taskList := make([]task, 0, len(inputs)*len(endpoints)*len(qtypes))
	tasks := make(chan task)
	results := make(chan Result)

	for _, input := range inputs {
		name, err := domain.Normalize(input, domain.Options{Strict: strict})
		if err != nil {
			all = append(all, Result{
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

	workerCount := concurrency
	if len(taskList) < workerCount {
		workerCount = len(taskList)
	}
	if workerCount == 0 {
		close(tasks)
		close(results)
		return all
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for t := range tasks {
				results <- resolveOne(ctx, t, opts)
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

func resolveOne(ctx context.Context, t task, opts resolver.Options) Result {
	qtypeName := resolver.QueryTypeName(t.qtype)
	base := Result{
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

func summarize(results []Result, duration time.Duration) Summary {
	summary := Summary{Total: len(results), DurationMS: int(duration.Milliseconds())}
	for _, res := range results {
		switch res.Status {
		case "OK":
			summary.OK++
		case "NO_ANSWER":
			summary.NoAnswer++
		case "DNS_ERROR":
			summary.DNSError++
		case "INVALID_DOMAIN":
			summary.Invalid++
		default:
			summary.Error++
		}
		if len(res.IPs) > 0 {
			summary.WithIPs++
		}
	}
	return summary
}

func splitInputs(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	return fields
}

func uniqueStrings(items []string) []string {
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
	return unique
}

func marshal(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return `{"error":"JSON 序列化失败"}`
	}
	return string(data)
}
