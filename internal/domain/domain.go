package domain

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

// Name is the normalized representation of a user-provided DNS name.
type Name struct {
	Original string   `json:"original"`
	Host     string   `json:"host"`
	ASCII    string   `json:"ascii"`
	FQDN     string   `json:"fqdn"`
	Root     bool     `json:"root"`
	Warnings []string `json:"warnings,omitempty"`
}

type Options struct {
	Strict bool
}

func Normalize(input string, opts Options) (Name, error) {
	original := input
	host, err := extractHost(input)
	if err != nil {
		return Name{Original: original}, err
	}

	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	if host == "" {
		return Name{Original: original}, errors.New("域名为空")
	}
	if ip := net.ParseIP(host); ip != nil {
		return Name{Original: original, Host: host}, fmt.Errorf("输入是 IP 地址，不是可查询的域名: %s", host)
	}
	if host == "." {
		return Name{
			Original: original,
			Host:     ".",
			ASCII:    ".",
			FQDN:     ".",
			Root:     true,
		}, nil
	}

	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return Name{Original: original}, errors.New("域名为空")
	}
	if strings.Contains(host, "..") {
		return Name{Original: original, Host: host}, fmt.Errorf("域名包含连续的点，DNS 不允许空 label: %s", host)
	}

	labels := strings.Split(host, ".")
	asciiLabels := make([]string, 0, len(labels))
	warnings := make([]string, 0)
	profile := idna.Lookup

	for _, label := range labels {
		if label == "" {
			return Name{Original: original, Host: host}, fmt.Errorf("域名包含空 label: %s", host)
		}
		if hasSpaceOrControl(label) {
			return Name{Original: original, Host: host}, fmt.Errorf("域名 label 包含空白或控制字符: %q", label)
		}

		ascii := label
		if hasNonASCII(label) {
			ascii, err = profile.ToASCII(label)
			if err != nil {
				return Name{Original: original, Host: host}, fmt.Errorf("IDN label 转换失败 %q: %w", label, err)
			}
		}
		ascii = strings.ToLower(ascii)

		if len(ascii) > 63 {
			return Name{Original: original, Host: host}, fmt.Errorf("域名 label 超过 63 字节: %q", ascii)
		}
		if err := validateASCII(label, ascii, opts.Strict); err != nil {
			return Name{Original: original, Host: host}, err
		}
		if strings.Contains(ascii, "_") {
			warnings = append(warnings, fmt.Sprintf("label %q 包含下划线；DNS 可查询，但它不是标准主机名 label", label))
		}
		if ascii == "*" {
			warnings = append(warnings, "检测到通配符 label *；将按字面 DNS 名称查询")
		}
		if strings.HasPrefix(ascii, "-") || strings.HasSuffix(ascii, "-") {
			warnings = append(warnings, fmt.Sprintf("label %q 以连字符开头或结尾；DNS 可查询，但它不是标准主机名 label", label))
		}

		asciiLabels = append(asciiLabels, ascii)
	}

	asciiName := strings.Join(asciiLabels, ".")
	if len(asciiName) > 253 {
		return Name{Original: original, Host: host}, fmt.Errorf("完整域名超过 253 字符: %d", len(asciiName))
	}

	return Name{
		Original: original,
		Host:     host,
		ASCII:    asciiName,
		FQDN:     asciiName + ".",
		Warnings: warnings,
	}, nil
}

func extractHost(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", errors.New("输入为空")
	}
	if s == "." {
		return ".", nil
	}

	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("URL 解析失败: %w", err)
		}
		host := u.Hostname()
		if host == "" {
			return "", fmt.Errorf("URL 中没有 hostname: %s", s)
		}
		return host, nil
	}

	if strings.ContainsAny(s, "/?#") {
		u, err := url.Parse("//" + s)
		if err == nil && u.Hostname() != "" {
			return u.Hostname(), nil
		}
	}

	if host, port, err := net.SplitHostPort(s); err == nil {
		if port != "" {
			if _, convErr := strconv.Atoi(port); convErr == nil {
				return host, nil
			}
		}
	}

	if idx := strings.LastIndex(s, ":"); idx > -1 && strings.Count(s, ":") == 1 {
		if _, err := strconv.Atoi(s[idx+1:]); err == nil {
			return s[:idx], nil
		}
	}

	return s, nil
}

func validateASCII(originalLabel, ascii string, strict bool) error {
	if ascii == "" {
		return fmt.Errorf("域名包含空 label: %q", originalLabel)
	}
	if ascii == "*" {
		return nil
	}
	for _, r := range ascii {
		ok := r == '-' || r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9')
		if !ok {
			return fmt.Errorf("域名 label 包含不支持的字符 %q: %q", r, originalLabel)
		}
	}
	if strict {
		for _, r := range ascii {
			ok := r == '-' ||
				(r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9')
			if !ok {
				return fmt.Errorf("严格模式下 label 只能包含字母、数字和连字符: %q", originalLabel)
			}
		}
		if strings.HasPrefix(ascii, "-") || strings.HasSuffix(ascii, "-") {
			return fmt.Errorf("严格模式下 label 不能以连字符开头或结尾: %q", originalLabel)
		}
	}
	return nil
}

func hasNonASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

func hasSpaceOrControl(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return true
		}
	}
	return false
}
