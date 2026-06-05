package main

import (
	"os"
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
