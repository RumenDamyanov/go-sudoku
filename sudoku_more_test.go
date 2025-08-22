package sudoku

import "testing"

func TestValidate(t *testing.T) {
	var b Board
	// place duplicates in a row
	b[0][0], b[0][1] = 5, 5
	if err := Validate(b); err == nil {
		t.Fatalf("expected invalid board due to row duplicate")
	}
	// fix row, put duplicate in column
	b[0][1] = 0
	b[0][0], b[1][0] = 7, 7
	if err := Validate(b); err == nil {
		t.Fatalf("expected invalid board due to column duplicate")
	}
	// fix column, put duplicate in box
	b[1][0] = 0
	b[0][0], b[1][1] = 3, 3
	if err := Validate(b); err == nil {
		t.Fatalf("expected invalid board due to box duplicate")
	}
}

func TestSolveUnsolvable(t *testing.T) {
	// Construct a board where the very first empty cell (0,0) has no legal moves.
	// Row 0 has 1..8 already, column 0 has 9.
	var b Board
	for c := 1; c < 9; c++ {
		b[0][c] = c // 1..8 across row 0
	}
	b[1][0] = 9 // blocks 9 in column 0
	if _, ok := Solve(b); ok {
		t.Fatalf("expected not solvable")
	}
}

func TestGenerateClueCounts(t *testing.T) {
	for _, d := range []Difficulty{Easy, Medium, Hard} {
		b, err := Generate(d, 1)
		if err != nil {
			t.Fatalf("generate %v: %v", d, err)
		}
		if err := Validate(b); err != nil {
			t.Fatalf("generated invalid board: %v", err)
		}
		// ensure unique
		if !hasUniqueSolution(b, 2) {
			t.Fatalf("generated board not unique for %v", d)
		}
	}
}

func TestHint(t *testing.T) {
	in := "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	b, err := FromString(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	r, c, v, ok := Hint(b)
	if !ok {
		t.Fatalf("expected a hint")
	}
	if r < 0 || r >= 9 || c < 0 || c >= 9 || v < 1 || v > 9 {
		t.Fatalf("bad hint: r=%d c=%d v=%d", r, c, v)
	}
}
