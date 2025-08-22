package examples

import (
	"fmt"

	"go.rumenx.com/sudoku"
)

// Example showing variable-sized grid.
func Example_gridGenerateHint() {
	sudoku.SetRandSeed(42)
	g, _ := sudoku.NewGrid(6, 2, 3)
	p, _ := g.Generate(sudoku.Easy, 2)
	r, c, v, ok := sudoku.HintGrid(p)
	fmt.Println("hint-ok:", ok, "cell:", r, c, "val:", v)
	// Output:
	// hint-ok: true cell: 0 1 val: 1
}
