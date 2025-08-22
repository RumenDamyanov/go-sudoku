# go-sudoku

[![CI](https://github.com/rumendamyanov/go-sudoku/actions/workflows/ci.yml/badge.svg)](https://github.com/rumendamyanov/go-sudoku/actions/workflows/ci.yml)
[![CodeQL](https://github.com/rumendamyanov/go-sudoku/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/rumendamyanov/go-sudoku/actions/workflows/github-code-scanning/codeql)
[![Dependabot](https://github.com/rumendamyanov/go-sudoku/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/rumendamyanov/go-sudoku/actions/workflows/dependabot/dependabot-updates)
[![codecov](https://codecov.io/gh/rumendamyanov/go-sudoku/graph/badge.svg)](https://codecov.io/gh/rumendamyanov/go-sudoku)
[![Go Report Card](https://goreportcard.com/badge/go.rumenx.com/sudoku?5)](https://goreportcard.com/report/go.rumenx.com/sudoku)
[![Go Reference](https://pkg.go.dev/badge/go.rumenx.com/sudoku.svg)](https://pkg.go.dev/go.rumenx.com/sudoku)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/rumendamyanov/go-sudoku/blob/master/LICENSE.md)

Fast, dependency-light Sudoku generator & solver for Go — classic 9x9 plus configurable smaller grids — with guaranteed unique puzzles, deterministic seeding, CLI, REST server, and optional GUI.

## Highlights

- ✅ Dependency-light core (stdlib only)
- 🎯 Deterministic, validated backtracking solver
- 🧩 Generator guarantees a unique solution & target clue counts (difficulty aware)
- 🧪 >90% core test coverage & CLI / server exercised
- 🌐 Minimal REST server (generate / solve)
- 💻 CLI (generate / solve / hint / variable sizes) & optional Fyne GUI (`-tags gui`)
- 🧱 General `Grid` API for 4x4, 6x6, 9x9 (and other <=9 sizes with box layout)
- 🔍 Hint helpers for both classic `Board` and general `Grid`

## Install

```sh
go get go.rumenx.com/sudoku@latest

import (
	"fmt"
	"go.rumenx.com/sudoku"
)
func main() {
	puz, _ := sudoku.Generate(sudoku.Medium, 1)
	fmt.Println("Puzzle (string):", puz.String())
	if sol, ok := sudoku.Solve(puz); ok {
	fmt.Println("Solved:", sol.String())
	}
}
import (
	"fmt"
	"go.rumenx.com/sudoku"
)
func main() {
	puz, _ := sudoku.Generate(sudoku.Medium, 1)
curl -s localhost:8080/health
	if sol, ok := sudoku.Solve(puz); ok {
		fmt.Println("Solved:", sol.String())
	}
}
```

## Variable Size (Grid API)

```go
g, _ := sudoku.NewGrid(6, 2, 3)          // 6x6 board with 2x3 sub-boxes
puz6, _ := g.Generate(sudoku.Easy, 1)     // unique puzzle
row, col, val, ok := sudoku.HintGrid(puz6)
_ = row; _ = col; _ = val; _ = ok
solved6, _ := puz6.Solve()
```

## Difficulty & Attempts

`Generate(difficulty, attempts)` uses `attempts` as a retry budget when searching for a satisfactorily carved unique puzzle (useful for harder settings). Set `attempts` > 1 if you want the generator to try again with fresh solved bases before returning.

| Difficulty | Target clue heuristic (9x9) |
|------------|------------------------------|
| Easy       | ~40 clues                    |
| Medium     | ~32 clues                    |
| Hard       | ~26 clues                    |

Other sizes scale proportionally.

## Library Cheat Sheet

Classic (9x9):

```go
b, _ := sudoku.FromString("53..7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79")
sol, ok := sudoku.Solve(b)
_ = sol; _ = ok
r,c,v,hk := sudoku.Hint(b)
_ = r; _ = c; _ = v; _ = hk
```

Generalized:

```go
g, _ := sudoku.NewGrid(4,2,2)
gp, _ := g.Generate(sudoku.Medium, 1)
gs, _ := gp.Solve()
_,_,_,_ = sudoku.HintGrid(gp)
_ = gs
```

## REST Server

Run:

```sh
make run      # or: go run ./cmd/server
```

Listens on `:8080` (override with `PORT`).

### Endpoints

| Method | Path      | Purpose                                      |
|--------|-----------|----------------------------------------------|
| GET    | /health   | Liveness & version (alias: /healthz)         |
| POST   | /generate | Generate puzzle (classic or variable size)   |
| POST   | /solve    | Solve or hint (classic or grid)              |

### POST /generate body

```jsonc
{
	"difficulty": "easy|medium|hard",    // optional (default medium)
	"includeSolution": true,              // optional (classic only adds solution)
	"size": 9,                            // optional (4,6,9) defaults 9 (classic path if omitted)
	"box": "3x3",                         // required whenever size provided (e.g. 2x2,2x3,3x3)
	"attempts": 3                         // optional retry budget for uniqueness
}
```

Response (classic or generalized) always returns a 2D numeric array for `puzzle` (and optional `solution`).

### Example Requests

```sh
# Generate hard 9x9 including solution
curl -s -X POST localhost:8080/generate \
	-H 'content-type: application/json' \
	-d '{"difficulty":"hard","includeSolution":true}' | jq '.puzzle | length'

# Generate 6x6 puzzle
curl -s -X POST localhost:8080/generate \
	-H 'content-type: application/json' \
	-d '{"size":6,"box":"2x3","difficulty":"easy"}' | jq '.puzzle | length'

# Solve a classic puzzle
curl -s -X POST localhost:8080/solve \
	-H 'content-type: application/json' \
	-d '{"string":"530070000600195000098000060800060003400803001700020006060000280000419005000080079"}' | jq '.solution'
```

## CLI

Build:

```sh
make build
./bin/sudoku-cli -version
```

Flags (subset):

| Flag        | Description                             |
|-------------|-----------------------------------------|
| -difficulty | easy / medium / hard (generate)         |
| -attempts   | retry budget for generation uniqueness  |
| -solve      | When generating also print solution     |
| -size       | Grid size (4,6,9) for generation        |
| -box        | Box dims RxC (2x2,2x3,3x3)              |
| -string     | Provide puzzle string to solve / hint   |
| -file       | File containing puzzle string           |
| -hint       | Print single hint (with -string/-file)  |
| -json       | JSON output                             |
| -version    | Print version and exit                  |

Examples:

```sh
# Generate medium 9x9 and show solution
./bin/sudoku-cli -difficulty medium -solve

# Generate 6x6
./bin/sudoku-cli -size 6 -box 2x3 -difficulty easy

# Solve string (JSON output)
./bin/sudoku-cli -string "530070000600195000098000060800060003400803001700020006060000280000419005000080079" -json

# Hint only
./bin/sudoku-cli -string "530070000600195000098000060800060003400803001700020006060000280000419005000080079" -hint
```

## GUI (Optional, Build Tag `gui`)

See `cmd/gui`. Build / run:

```sh
make gui          # run with tag
make build-gui    # build binary ./bin/sudoku-gui
```

Features: size selector (4/6/9), difficulty, timer, hint, validate, solve, clear, theme styling.

<!-- Removed duplicate Docker heading earlier in document -->
Classic:

```go
type Board [9][9]int
func Validate(Board) error
func Solve(Board) (Board, bool)
func Generate(Difficulty, int) (Board, error)
func FromString(string) (Board, error)
func (Board) String() string
func Hint(Board) (row, col, val int, ok bool)
```

Generalized:

```go
type Grid struct { Size, BoxRows, BoxCols int; Cells [][]int }
func NewGrid(size, boxRows, boxCols int) (Grid, error)
func (Grid) Validate() error
func (Grid) Solve() (Grid, bool)
func (Grid) Generate(Difficulty, int) (Grid, error)
func FromStringN(s string, size, boxRows, boxCols int) (Grid, error)
func (Grid) String() string
func HintGrid(Grid) (row, col, val int, ok bool)
```

## Acknowledgements

Backtracking solver pattern adapted for clarity & determinism. All code written from scratch for this project.

## Docker

Build the image:

```sh
make docker-build
# or
docker build -t go-sudoku:local .
```

Run the server in Docker:

```sh
make docker-run
# or
docker run --rm -p 8080:8080 -e PORT=8080 go-sudoku:local
```

Then try the API:

```sh
curl -s localhost:8080/health
curl -s -X POST localhost:8080/generate -H 'content-type: application/json' -d '{"difficulty":"medium"}'
```

## Community & Security

See: [CONTRIBUTING](CONTRIBUTING.md) · [CODE_OF_CONDUCT](CODE_OF_CONDUCT.md) · [SECURITY](SECURITY.md) · [FUNDING](FUNDING.md)

Report vulnerabilities per SECURITY guidelines. Contributions should follow the code of conduct.

### Module usage and examples

The library exposes a simple API around a 9x9 `Board`.

- Generate a unique puzzle:

```go
puz, err := sudoku.Generate(sudoku.Medium, 1)
if err != nil { /* handle */ }

// Persist or send over the wire as 81-char string:
str := puz.String() // "...81 chars..."

// Parse back later:
back, err := sudoku.FromString(str)
```

- Solve a puzzle and validate:

```
An optional desktop GUI demo is included under `cmd/gui` using Fyne. It’s behind a build tag (`gui`) so the core module remains stdlib-only by default.

Requirements:

- Go 1.22+
- Fyne toolkit: `go get fyne.io/fyne/v2`

Run the GUI:

```sh
# install dependency (once, in this repo)
go get fyne.io/fyne/v2

# run with the gui build tag
make gui
# or build a binary
make build-gui
./bin/sudoku-gui
```

Features:

- Variable board sizes: 4x4 (2x2), 6x6 (2x3), 9x9 (3x3)
- Difficulty selector (easy/medium/hard)
- Generate, Solve, Validate, Clear
- Hint button: select a cell and click “Hint” to fill a valid value
- Timer: shows time since last generation
- Modern look: subtle box shading and focused-cell highlight

Troubleshooting:

- If you see missing go.sum entries, re-run the command; the Makefile uses `-mod=mod` to auto-resolve them.
- On macOS, GUI linking may emit harmless duplicate library warnings.
- If `go get fyne.io/fyne/v2` fails, ensure Xcode Command Line Tools are installed.

Extending the GUI:

- This is a small demo meant to be extended. Core state is in `gridState`, and the library exposes general `Grid` APIs: `NewGrid`, `Grid.Generate`, `Grid.Solve`, `Grid.Validate`, `HintGrid`.
- Ideas: pencil marks, undo/redo, mistake highlights, keyboard navigation, themes, persistence, larger sizes (e.g., 12x12 with 3x4 boxes).
- For bigger apps, extract grid widgets/state into a separate package and compose more views on top.

## License

MIT — see [LICENSE.md](LICENSE.md).
