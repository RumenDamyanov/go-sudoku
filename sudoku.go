// Package sudoku provides a simple Sudoku board representation, solver, and generator.
// Module: go.rumenx.com/sudoku
// Author: Rumen Damyanov <contact@rumenx.com>
package sudoku

import (
	"errors"
	"math/rand/v2"
)

// Board is a 9x9 Sudoku grid. Empty cells are 0.
type Board [9][9]int

// Difficulty controls the number of clues in a generated puzzle.
type Difficulty string

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

var (
	// ErrInvalidBoard is returned when a board violates Sudoku rules.
	ErrInvalidBoard = errors.New("invalid board")
	// globalRand is the random source used by generator; overridden via SetRandSeed.
	globalRand = rand.New(rand.NewPCG(uint64(rand.Uint32()), uint64(rand.Uint32())))
)

// SetRandSeed sets the seed for the library's random generator ensuring reproducible generation.
// Safe for tests; not concurrency guarded (call during init).
func SetRandSeed(seed uint64) { globalRand = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)) }

// Validate checks that values are in [0,9] and no row/col/box duplicates (ignoring zeros).
func Validate(b Board) error {
	// rows and cols
	for i := 0; i < 9; i++ {
		var row, col [10]bool
		for j := 0; j < 9; j++ {
			rv := b[i][j]
			cv := b[j][i]
			if rv < 0 || rv > 9 || cv < 0 || cv > 9 {
				return ErrInvalidBoard
			}
			if rv != 0 {
				if row[rv] {
					return ErrInvalidBoard
				}
				row[rv] = true
			}
			if cv != 0 {
				if col[cv] {
					return ErrInvalidBoard
				}
				col[cv] = true
			}
		}
	}
	// 3x3 boxes
	for br := 0; br < 9; br += 3 {
		for bc := 0; bc < 9; bc += 3 {
			var seen [10]bool
			for r := br; r < br+3; r++ {
				for c := bc; c < bc+3; c++ {
					v := b[r][c]
					if v != 0 {
						if seen[v] {
							return ErrInvalidBoard
						}
						seen[v] = true
					}
				}
			}
		}
	}
	return nil
}

// Solve tries to solve the board using backtracking. Returns solved board and ok.
func Solve(b Board) (Board, bool) {
	var solved Board
	copyBoard(&solved, &b)
	if !backtrack(&solved) {
		return Board{}, false
	}
	return solved, true
}

// backtrack fills empty cells; standard DFS.
func backtrack(b *Board) bool {
	r, c, ok := findEmpty(b)
	if !ok {
		return true
	}
	// try 1..9 shuffled for some variety
	vals := [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	globalRand.Shuffle(9, func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })
	for _, v := range vals {
		if isSafe(*b, r, c, v) {
			b[r][c] = v
			if backtrack(b) {
				return true
			}
			b[r][c] = 0
		}
	}
	return false
}

func findEmpty(b *Board) (int, int, bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

func isSafe(b Board, r, c, v int) bool {
	for i := 0; i < 9; i++ {
		if b[r][i] == v || b[i][c] == v {
			return false
		}
	}
	br := (r / 3) * 3
	bc := (c / 3) * 3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b[br+i][bc+j] == v {
				return false
			}
		}
	}
	return true
}

func copyBoard(dst, src *Board) {
	for i := range dst {
		dst[i] = src[i]
	}
}

// Generate creates a Sudoku puzzle with a unique solution.
// attempts controls how many removal passes to try; set to >= 1.
func Generate(d Difficulty, attempts int) (Board, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for try := 0; try < attempts; try++ {
		var b Board
		fillDiagonalBoxes(&b)
		if !backtrack(&b) {
			lastErr = errors.New("failed to build solved board")
			continue
		}
		solution := b
		target := cluesFor(d)
		puzzle := solution
		rmOrder := globalRand.Perm(81)
		for _, idx := range rmOrder {
			if countClues(puzzle) <= target {
				break
			}
			r := idx / 9
			c := idx % 9
			old := puzzle[r][c]
			if old == 0 {
				continue
			}
			puzzle[r][c] = 0
			if !hasUniqueSolution(puzzle, 2) {
				puzzle[r][c] = old
			}
		}
		if hasUniqueSolution(puzzle, 2) { // uniqueness sanity
			return puzzle, nil
		}
		lastErr = errors.New("puzzle uniqueness not achieved")
	}
	if lastErr == nil {
		lastErr = errors.New("generation failed")
	}
	return Board{}, lastErr
}

func cluesFor(d Difficulty) int {
	switch d {
	case Easy:
		return 40
	case Medium:
		return 32
	case Hard:
		return 26
	default:
		return 32
	}
}

func countClues(b Board) int {
	cnt := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != 0 {
				cnt++
			}
		}
	}
	return cnt
}

// hasUniqueSolution returns true if the board has exactly one solution, early stopping after 'limit' found.
func hasUniqueSolution(b Board, limit int) bool {
	count := 0
	var work Board
	copyBoard(&work, &b)
	var dfs func(*Board) bool
	dfs = func(cur *Board) bool {
		r, c, ok := findEmpty(cur)
		if !ok {
			count++
			return count >= limit // early-exit if we hit the limit
		}
		for v := 1; v <= 9; v++ {
			if isSafe(*cur, r, c, v) {
				cur[r][c] = v
				if dfs(cur) { // early exit propagate
					return true
				}
				cur[r][c] = 0
			}
		}
		return false
	}
	dfs(&work)
	return count == 1
}

func fillDiagonalBoxes(b *Board) {
	for d := 0; d < 9; d += 3 {
		fillBox(b, d, d)
	}
}

func fillBox(b *Board, br, bc int) {
	vals := globalRand.Perm(9)
	idx := 0
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			b[br+r][bc+c] = vals[idx] + 1
			idx++
		}
	}
}
