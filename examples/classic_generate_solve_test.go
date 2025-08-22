package examples

import (
	"fmt"

	"go.rumenx.com/sudoku"
)

// Example demonstrating classic 9x9 generation and solving.
func Example_classicGenerateSolve() {
	sudoku.SetRandSeed(1)
	puz, _ := sudoku.Generate(sudoku.Medium, 2)
	prefix := puz.String()[:9]
	fmt.Println("clues:", prefix)
	if sol, ok := sudoku.Solve(puz); ok {
		_ = sol
		fmt.Println("solvable: true")
	}
	// Output:
	// clues: 400705000
	// solvable: true
}
