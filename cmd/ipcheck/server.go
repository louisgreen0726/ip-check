package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"ipcheck/internal/endpoint"
	"ipcheck/internal/resolver"
)

//go:embed web/index.html web/app.css web/app.js
var webAssets embed.FS

type serveOptions struct {
	addr string
	open bool
}

type resolveRequest struct {
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
	IPInfo             *bool    `json:"ipInfo"`
}

type resolveResponse struct {
	Results []result       `json:"results"`
	Summary resolveSummary `json:"summary"`
}

type resolveSummary struct {
	Total      int `json:"total"`
	OK         int `json:"ok"`
	NoAnswer   int `json:"noAnswer"`
	DNSError   int `json:"dnsError"`
	Error      int `json:"error"`
	Invalid    int `json:"invalid"`
	WithIPs    int `json:"withIps"`
	DurationMS int `json:"durationMs"`
}

func runServer(args []string, stdout, stderr io.Writer) error {
	opts := serveOptions{}
	fs := flag.NewFlagSet("ipcheck serve", flag.ExitOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&opts.addr, "addr", "127.0.0.1:8765", "GUI 监听地址")
	fs.BoolVar(&opts.open, "open", false, "启动后打开浏览器")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: ipcheck serve [options]\n\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	mux := newServerMux()
	server := &http.Server{
		Addr:              opts.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	url := displayURL(opts.addr)
	fmt.Fprintf(stdout, "IP Check GUI listening on %s\n", url)
	if opts.open {
		go openBrowser(url)
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func newServerMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveAsset)
	mux.HandleFunc("/api/resolve", handleResolveAPI)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
	})
	return mux
}

func serveAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	if strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}

	data, err := webAssets.ReadFile("web/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

func handleResolveAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	defer r.Body.Close()

	var req resolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "JSON 请求解析失败: " + err.Error()})
		return
	}

	start := time.Now()
	results, err := resolveFromRequest(r.Context(), req)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	resp := resolveResponse{
		Results: results,
		Summary: summarizeResults(results, time.Since(start)),
	}
	writeJSONResponse(w, http.StatusOK, resp)
}

func resolveFromRequest(ctx context.Context, req resolveRequest) ([]result, error) {
	domains := make([]string, 0, len(req.Domains))
	domains = append(domains, req.Domains...)
	domains = append(domains, splitInputs(req.DomainText)...)
	domains = uniqueStrings(domains)
	if len(domains) == 0 {
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
	dohMethod := strings.ToUpper(strings.TrimSpace(req.DoHMethod))
	if dohMethod == "" {
		dohMethod = "POST"
	}

	ednsEnabled := true
	if req.EDNS != nil {
		ednsEnabled = *req.EDNS
	}

	cli := cliOptions{
		timeout:            timeout,
		retries:            req.Retries,
		concurrency:        concurrency,
		strict:             req.Strict,
		noEDNS:             !ednsEnabled,
		dnssec:             req.DNSSEC,
		dohMethod:          dohMethod,
		insecureSkipVerify: req.InsecureSkipVerify,
	}

	results := resolveAll(ctx, domains, endpoints, qtypes, cli)
	if req.IPInfo != nil && *req.IPInfo {
		enrichIPInfo(ctx, results)
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
	return results, nil
}

func summarizeResults(results []result, duration time.Duration) resolveSummary {
	summary := resolveSummary{Total: len(results), DurationMS: int(duration.Milliseconds())}
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

func writeJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func displayURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0:") {
		return "http://127.0.0.1:" + strings.TrimPrefix(addr, "0.0.0.0:")
	}
	return "http://" + addr
}

func openBrowser(rawURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	_ = cmd.Start()
}
