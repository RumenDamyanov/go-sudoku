package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	out := &bytes.Buffer{}
	code := runCLI([]string{"-version"}, out, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("expected exit 0 got %d", code)
	}
	re := regexp.MustCompile(`^sudoku-cli \S+ \(commit \S+, built \S+\)$`)
	s := strings.TrimSpace(out.String())
	if !re.MatchString(s) {
		t.Fatalf("unexpected version output: %q", s)
	}
}
