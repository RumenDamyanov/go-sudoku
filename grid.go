package sudoku

import (
	"errors"
	"fmt"
)

// Maximum allowed grid size to prevent excessive memory usage.
const MaxGridSize = 25

// Grid is a generalised Sudoku grid of size SxS with sub-boxes boxRows x boxCols,
// where S == boxRows*boxCols. Values are in [0..S], 0 meaning empty.
type Grid struct {
	Size    int
	BoxRows int
	BoxCols int
	Cells   [][]int // length Size, each length Size
}

// NewGrid creates an empty grid with given dimensions.
func NewGrid(size, boxRows, boxCols int) (Grid, error) {
	if size <= 0 || boxRows <= 0 || boxCols <= 0 || size != boxRows*boxCols {
		return Grid{}, fmt.Errorf("invalid dimensions: size=%d boxRows=%d boxCols=%d", size, boxRows, boxCols)
	}
	if size > MaxGridSize {
		return Grid{}, fmt.Errorf("grid size %d exceeds maximum allowed (%d)", size, MaxGridSize)
	}
	g := Grid{Size: size, BoxRows: boxRows, BoxCols: boxCols, Cells: make([][]int, size)}
	for i := range g.Cells {
		g.Cells[i] = make([]int, size)
	}
	return g, nil
}

// Clone returns a deep copy of the grid.
func (g Grid) Clone() Grid {
	out, _ := NewGrid(g.Size, g.BoxRows, g.BoxCols)
	for r := 0; r < g.Size; r++ {
		copy(out.Cells[r], g.Cells[r])
	}
	return out
}

// Validate checks that values are in [0..Size] and no row/col/box duplicates (ignoring zeros).
func (g Grid) Validate() error {
	s := g.Size
	// rows and cols
	for i := 0; i < s; i++ {
		row := make([]bool, s+1)
		col := make([]bool, s+1)
		for j := 0; j < s; j++ {
			rv := g.Cells[i][j]
			cv := g.Cells[j][i]
			if rv < 0 || rv > s || cv < 0 || cv > s {
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
	// boxes
	for br := 0; br < s; br += g.BoxRows {
		for bc := 0; bc < s; bc += g.BoxCols {
			seen := make([]bool, s+1)
			for r := br; r < br+g.BoxRows; r++ {
				for c := bc; c < bc+g.BoxCols; c++ {
					v := g.Cells[r][c]
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

// Solve tries to solve the grid using backtracking. Returns solved grid and ok.
func (g Grid) Solve() (Grid, bool) {
	work := g.Clone()
	if !g.backtrack(&work) {
		return Grid{}, false
	}
	return work, true
}

func (g Grid) backtrack(w *Grid) bool {
	r, c, ok := g.findEmpty(w)
	if !ok {
		return true
	}
	// try values 1..Size shuffled for variety
	vals := make([]int, g.Size)
	for i := 0; i < g.Size; i++ {
		vals[i] = i + 1
	}
	globalRand.Shuffle(len(vals), func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })
	for _, v := range vals {
		if g.isSafe(*w, r, c, v) {
			w.Cells[r][c] = v
			if g.backtrack(w) {
				return true
			}
			w.Cells[r][c] = 0
		}
	}
	return false
}

func (g Grid) findEmpty(w *Grid) (int, int, bool) {
	for r := 0; r < g.Size; r++ {
		for c := 0; c < g.Size; c++ {
			if w.Cells[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

func (g Grid) isSafe(w Grid, r, c, v int) bool {
	for i := 0; i < g.Size; i++ {
		if w.Cells[r][i] == v || w.Cells[i][c] == v {
			return false
		}
	}
	br := (r / g.BoxRows) * g.BoxRows
	bc := (c / g.BoxCols) * g.BoxCols
	for i := 0; i < g.BoxRows; i++ {
		for j := 0; j < g.BoxCols; j++ {
			if w.Cells[br+i][bc+j] == v {
				return false
			}
		}
	}
	return true
}

// Generate creates a puzzle with a unique solution.
func (g Grid) Generate(d Difficulty, attempts int) (Grid, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for try := 0; try < attempts; try++ {
		solved := g.Clone()
		solved.fillDiagonalBoxes()
		if !g.backtrack(&solved) {
			lastErr = errors.New("failed to build solved grid")
			continue
		}
		target := g.cluesFor(d)
		puzzle := solved.Clone()
		rmOrder := globalRand.Perm(g.Size * g.Size)
		for _, idx := range rmOrder {
			if g.countClues(puzzle) <= target {
				break
			}
			r := idx / g.Size
			c := idx % g.Size
			old := puzzle.Cells[r][c]
			if old == 0 {
				continue
			}
			puzzle.Cells[r][c] = 0
			if !g.hasUniqueSolution(puzzle, 2) {
				puzzle.Cells[r][c] = old
			}
		}
		if g.hasUniqueSolution(puzzle, 2) {
			return puzzle, nil
		}
		lastErr = errors.New("puzzle uniqueness not achieved")
	}
	if lastErr == nil {
		lastErr = errors.New("generation failed")
	}
	return Grid{}, lastErr
}

func (g Grid) cluesFor(d Difficulty) int {
	// Scale classic clue counts by size ratio (baseline 9x9)
	base := 32 // medium
	switch d {
	case Easy:
		base = 40
	case Hard:
		base = 26
	}
	if g.Size == 9 {
		return base
	}
	// proportionally adjust number of clues relative to 9x9 (81 cells)
	cells := g.Size * g.Size
	return max(8, cells*base/81) // keep a minimal clue count
}

func (g Grid) countClues(w Grid) int {
	cnt := 0
	for r := 0; r < g.Size; r++ {
		for c := 0; c < g.Size; c++ {
			if w.Cells[r][c] != 0 {
				cnt++
			}
		}
	}
	return cnt
}

// hasUniqueSolution returns true if there is exactly one solution, with early stop at limit.
func (g Grid) hasUniqueSolution(w Grid, limit int) bool {
	count := 0
	work := w.Clone()
	var dfs func(*Grid) bool
	dfs = func(cur *Grid) bool {
		r, c, ok := g.findEmpty(cur)
		if !ok {
			count++
			return count >= limit
		}
		for v := 1; v <= g.Size; v++ {
			if g.isSafe(*cur, r, c, v) {
				cur.Cells[r][c] = v
				if dfs(cur) {
					return true
				}
				cur.Cells[r][c] = 0
			}
		}
		return false
	}
	dfs(&work)
	return count == 1
}

func (g *Grid) fillDiagonalBoxes() {
	// For rectangular boxes, step across the diagonal in box coordinates.
	// Number of box rows and cols:
	nRowBoxes := g.Size / g.BoxRows
	nColBoxes := g.Size / g.BoxCols
	steps := nRowBoxes
	if nColBoxes < steps {
		steps = nColBoxes
	}
	for i := 0; i < steps; i++ {
		br := i * g.BoxRows
		bc := i * g.BoxCols
		g.fillBox(br, bc)
	}
}

func (g *Grid) fillBox(br, bc int) {
	vals := globalRand.Perm(g.Size)
	idx := 0
	for r := 0; r < g.BoxRows; r++ {
		for c := 0; c < g.BoxCols; c++ {
			g.Cells[br+r][bc+c] = vals[idx] + 1
			idx++
		}
	}
}

// FromStringN parses a size*size characters string into a Grid.
// Digits 1-9 are values; 0 or '.' are empty. Supports sizes up to 9.
func FromStringN(s string, size, boxRows, boxCols int) (Grid, error) {
	if size != boxRows*boxCols {
		return Grid{}, fmt.Errorf("invalid dims: size=%d boxRows=%d boxCols=%d", size, boxRows, boxCols)
	}
	expected := size * size
	if len(s) != expected {
		return Grid{}, fmt.Errorf("input must be %d characters", expected)
	}
	g, _ := NewGrid(size, boxRows, boxCols)
	for i := 0; i < expected; i++ {
		ch := s[i]
		r := i / size
		c := i % size
		switch ch {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			v := int(ch - '0')
			if v > size {
				return Grid{}, errors.New("digit exceeds grid size")
			}
			g.Cells[r][c] = v
		case '0', '.':
			g.Cells[r][c] = 0
		default:
			return Grid{}, errors.New("invalid character in grid")
		}
	}
	if err := g.Validate(); err != nil {
		return Grid{}, err
	}
	return g, nil
}

// String returns the compact representation of a Grid (size*size runes, 0 for empty).
func (g Grid) String() string {
	buf := make([]byte, 0, g.Size*g.Size)
	for r := 0; r < g.Size; r++ {
		for c := 0; c < g.Size; c++ {
			v := g.Cells[r][c]
			if v == 0 {
				buf = append(buf, '0')
			} else {
				buf = append(buf, byte('0'+v))
			}
		}
	}
	return string(buf)
}

// Hint returns a single suggested value for the provided 9x9 Board.
// It returns row, col, value and true if a hint is available.
func Hint(b Board) (int, int, int, bool) {
	if err := Validate(b); err != nil {
		return 0, 0, 0, false
	}
	if sol, ok := Solve(b); ok {
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if b[r][c] == 0 {
					return r, c, sol[r][c], true
				}
			}
		}
	}
	return 0, 0, 0, false
}

// HintGrid returns a suggested value for a general Grid, if solvable.
func HintGrid(g Grid) (int, int, int, bool) {
	if err := g.Validate(); err != nil {
		return 0, 0, 0, false
	}
	if sol, ok := g.Solve(); ok {
		for r := 0; r < g.Size; r++ {
			for c := 0; c < g.Size; c++ {
				if g.Cells[r][c] == 0 {
					return r, c, sol.Cells[r][c], true
				}
			}
		}
	}
	return 0, 0, 0, false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
