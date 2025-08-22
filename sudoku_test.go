package sudoku

import "testing"

func TestParseAndString(t *testing.T) {
	in := "53..7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79"
	b, err := FromString(in)
	if err != nil {
		t.Fatalf("FromString error: %v", err)
	}
	if got := b.String(); len(got) != 81 {
		t.Fatalf("String length = %d", len(got))
	}
}

func TestSolveSimple(t *testing.T) {
	in := "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	b, err := FromString(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	solved, ok := Solve(b)
	if !ok {
		t.Fatalf("failed to solve")
	}
	if err := Validate(solved); err != nil {
		t.Fatalf("invalid solution: %v", err)
	}
}

func TestGenerate(t *testing.T) {
	puz, err := Generate(Easy, 1)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if err := Validate(puz); err != nil {
		t.Fatalf("generated invalid puzzle: %v", err)
	}
	if !hasUniqueSolution(puz, 2) { // package-private helper
		t.Fatalf("puzzle not unique")
	}
}
