package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI_MainGenerateJSON(t *testing.T) {
	// Call main once with a happy path that does not call os.Exit.
	var outBuf, errBuf bytes.Buffer
	code := runCLI([]string{"-json", "-difficulty=easy", "-size=9", "-box=3x3"}, &outBuf, &errBuf)
	if code != 0 {
		t.Fatalf("exit code %d, stderr=%s", code, errBuf.String())
	}
	out := outBuf.String()
	if !strings.Contains(out, "\"puzzle\"") {
		t.Fatalf("expected puzzle json in output, got: %s", out)
	}
}

func TestCLI_MainSolveAndHintJSON(t *testing.T) {
	puzzle := "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	for _, args := range [][]string{
		{"-string", puzzle, "-json"},
		{"-string", puzzle, "-hint", "-json"},
	} {
		var outBuf, errBuf bytes.Buffer
		code := runCLI(args, &outBuf, &errBuf)
		if code != 0 {
			t.Fatalf("exit code %d, stderr=%s", code, errBuf.String())
		}
		out := outBuf.String()
		if strings.Contains(strings.Join(args, " "), "-hint") {
			if !strings.Contains(out, "\"row\"") {
				t.Fatalf("expected hint json, got: %s", out)
			}
		} else {
			if !strings.Contains(out, "\"solution\"") {
				t.Fatalf("expected solution json, got: %s", out)
			}
		}
	}
}

func TestCLI_BadFlagsAndUnsolvable(t *testing.T) {
	// invalid difficulty
	{
		var outBuf, errBuf bytes.Buffer
		code := runCLI([]string{"-json", "-difficulty=unknown", "-size=9", "-box=3x3"}, &outBuf, &errBuf)
		if code == 0 {
			t.Fatalf("expected non-zero exit for invalid difficulty")
		}
		if !strings.Contains(errBuf.String(), "invalid difficulty") {
			t.Fatalf("stderr should mention invalid difficulty, got: %s", errBuf.String())
		}
	}
	// invalid box dims (size mismatch)
	{
		var outBuf, errBuf bytes.Buffer
		code := runCLI([]string{"-json", "-difficulty=easy", "-size=6", "-box=3x3"}, &outBuf, &errBuf)
		if code == 0 {
			t.Fatalf("expected non-zero exit for invalid box dims")
		}
		if !strings.Contains(errBuf.String(), "invalid box dims") {
			t.Fatalf("stderr should mention invalid box dims, got: %s", errBuf.String())
		}
	}
	// unsolvable but valid puzzle string
	{
		// Row0 has 1..8, col0 has 9 -> no legal move at (0,0)
		puzzle := "012345678" + "900000000" + strings.Repeat("000000000", 7)
		var outBuf, errBuf bytes.Buffer
		code := runCLI([]string{"-string", puzzle, "-json"}, &outBuf, &errBuf)
		if code == 0 {
			t.Fatalf("expected non-zero exit for unsolvable puzzle")
		}
		if !strings.Contains(errBuf.String(), "unsolvable puzzle") {
			t.Fatalf("stderr should mention unsolvable, got: %s", errBuf.String())
		}
	}
}
