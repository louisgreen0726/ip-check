package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestShouldLaunchGUIByDefault(t *testing.T) {
	if !shouldLaunchGUIByDefault(nil, nil) {
		t.Fatal("empty args without stdin should launch GUI by default")
	}
	if shouldLaunchGUIByDefault([]string{"example.com"}, nil) {
		t.Fatal("domain arguments should keep CLI behavior")
	}
}

func TestShouldLaunchGUIByDefaultKeepsPipedStdinCLI(t *testing.T) {
	stdin, stdout, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer stdin.Close()
	defer stdout.Close()

	if shouldLaunchGUIByDefault(nil, stdin) {
		t.Fatal("piped stdin should keep CLI behavior")
	}
}

func TestRunReturnsFlagParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--definitely-not-a-flag"}, &stdout, &stderr, nil)
	if err == nil {
		t.Fatal("expected invalid flag error")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRejectsUnsafeRuntimeOptions(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--retries", "-1", "example.com"}, &stdout, &stderr, nil)
	if err == nil {
		t.Fatal("expected negative retries to be rejected")
	}
	if !strings.Contains(err.Error(), "--retries") {
		t.Fatalf("unexpected error: %v", err)
	}
}
