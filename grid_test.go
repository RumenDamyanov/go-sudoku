package sudoku

import "testing"

func TestNewGridErrors(t *testing.T) {
	if _, err := NewGrid(9, 2, 5); err == nil { // 2*5 != 9
		t.Fatalf("expected invalid dims error")
	}
}

func TestGridValidateAndSolve4x4(t *testing.T) {
	g, err := NewGrid(4, 2, 2)
	if err != nil {
		t.Fatalf("new grid: %v", err)
	}
	// Simple 4x4 puzzle
	// 0 0 3 4
	// 3 4 0 0
	// 0 0 4 3
	// 4 3 0 0
	g.Cells = [][]int{{0, 0, 3, 4}, {3, 4, 0, 0}, {0, 0, 4, 3}, {4, 3, 0, 0}}
	if err := g.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	sol, ok := g.Solve()
	if !ok {
		t.Fatalf("solve failed")
	}
	if err := sol.Validate(); err != nil {
		t.Fatalf("solution invalid: %v", err)
	}
}

func TestGridGenerateAndHint(t *testing.T) {
	for _, cfg := range []struct{ size, br, bc int }{{4, 2, 2}, {6, 2, 3}, {9, 3, 3}} {
		g, err := NewGrid(cfg.size, cfg.br, cfg.bc)
		if err != nil {
			t.Fatalf("new grid: %v", err)
		}
		// Some sizes may rarely fail on first attempt; retry a few times to avoid flakiness.
		var puz Grid
		var genErr error
		for tries := 0; tries < 3; tries++ {
			puz, genErr = g.Generate(Medium, 1)
			if genErr == nil {
				break
			}
		}
		if genErr != nil {
			t.Fatalf("generate: %v", genErr)
		}
		if err := puz.Validate(); err != nil {
			t.Fatalf("generated invalid: %v", err)
		}
		if !puz.hasUniqueSolution(puz, 2) { // method on Grid
			t.Fatalf("puzzle not unique for size %d", cfg.size)
		}
		r, c, v, ok := HintGrid(puz)
		if !ok || r < 0 || c < 0 || v <= 0 {
			t.Fatalf("expected a hint for size %d", cfg.size)
		}
	}
}

func TestFromStringN(t *testing.T) {
	// 4x4 string parse
	s := "0000340000430000"
	g, err := FromStringN(s, 4, 2, 2)
	if err != nil {
		t.Fatalf("parse 4x4: %v", err)
	}
	if len(g.Cells) != 4 || len(g.Cells[0]) != 4 {
		t.Fatalf("bad dims")
	}
	// size mismatch
	if _, err := FromStringN("0000", 6, 2, 3); err == nil {
		t.Fatalf("expected size error")
	}
	// invalid char
	if _, err := FromStringN("x000", 2, 1, 2); err == nil {
		t.Fatalf("expected invalid char error")
	}
}
