package domain

import (
	"strings"
	"testing"
)

func TestNormalizeDeepDomain(t *testing.T) {
	name, err := Normalize("a.b.c.d.e.f.g.h.i.j.example.com", Options{})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if name.ASCII != "a.b.c.d.e.f.g.h.i.j.example.com" {
		t.Fatalf("unexpected ASCII name: %s", name.ASCII)
	}
	if name.FQDN != "a.b.c.d.e.f.g.h.i.j.example.com." {
		t.Fatalf("unexpected FQDN: %s", name.FQDN)
	}
}

func TestNormalizeIDN(t *testing.T) {
	name, err := Normalize("例子.测试", Options{})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if name.ASCII != "xn--fsqu00a.xn--0zwm56d" {
		t.Fatalf("unexpected punycode: %s", name.ASCII)
	}
}

func TestNormalizeURLAndPort(t *testing.T) {
	name, err := Normalize("https://Example.COM:8443/a/path?x=1", Options{})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if name.ASCII != "example.com" {
		t.Fatalf("unexpected host extraction: %s", name.ASCII)
	}

	name, err = Normalize("Example.org:443", Options{})
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if name.ASCII != "example.org" {
		t.Fatalf("unexpected host: %s", name.ASCII)
	}
}

func TestNormalizeRejectsEmptyLabel(t *testing.T) {
	_, err := Normalize("a..b.com", Options{})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
	if !strings.Contains(err.Error(), "连续的点") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeLabelLength(t *testing.T) {
	valid := strings.Repeat("a", 63) + ".example.com"
	if _, err := Normalize(valid, Options{}); err != nil {
		t.Fatalf("63-byte label should be valid: %v", err)
	}

	invalid := strings.Repeat("a", 64) + ".example.com"
	if _, err := Normalize(invalid, Options{}); err == nil {
		t.Fatal("expected 64-byte label to be invalid")
	}
}

func TestNormalizeUnderscoreWarningAndStrictMode(t *testing.T) {
	name, err := Normalize("_sip._tcp.example.com", Options{})
	if err != nil {
		t.Fatalf("underscore labels should be accepted in lenient mode: %v", err)
	}
	if len(name.Warnings) == 0 {
		t.Fatal("expected underscore warning")
	}

	if _, err := Normalize("_sip._tcp.example.com", Options{Strict: true}); err == nil {
		t.Fatal("expected strict mode to reject underscore labels")
	}
}

func TestNormalizeRejectsUnsupportedCharacters(t *testing.T) {
	for _, input := range []string{
		"ex*ample.com",
		`bad\name.example.com`,
	} {
		if _, err := Normalize(input, Options{}); err == nil {
			t.Fatalf("expected %q to be rejected", input)
		}
	}
}

func TestNormalizeAllowsWildcardLabel(t *testing.T) {
	name, err := Normalize("*.example.com", Options{})
	if err != nil {
		t.Fatalf("wildcard label should be accepted: %v", err)
	}
	if name.ASCII != "*.example.com" {
		t.Fatalf("unexpected wildcard normalization: %s", name.ASCII)
	}
}

func TestNormalizeRoot(t *testing.T) {
	name, err := Normalize(".", Options{})
	if err != nil {
		t.Fatalf("root should be accepted: %v", err)
	}
	if !name.Root || name.FQDN != "." {
		t.Fatalf("unexpected root result: %+v", name)
	}
}
