package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"go.rumenx.com/sudoku"
)

var (
	// These can be overridden via -ldflags "-X main.version=... -X main.commit=... -X main.date=..."
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func versionString() string {
	return fmt.Sprintf("sudoku-cli %s (commit %s, built %s)", version, commit, date)
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

// runCLI executes the CLI with provided args and I/O, returning a process exit code.
func runCLI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("sudoku-cli", flag.ContinueOnError)
	fs.SetOutput(stderr)
	diff := fs.String("difficulty", "medium", "difficulty: easy|medium|hard (for generation)")
	attempts := fs.Int("attempts", 3, "generation attempts for uniqueness (>=1)")
	showSol := fs.Bool("solve", false, "when generating, also show solution")
	size := fs.Int("size", 9, "grid size (SxS), e.g. 4, 6, 9")
	box := fs.String("box", "3x3", "sub-box dims RxC, e.g. 2x2 for 4x4, 2x3 for 6x6, 3x3 for 9x9")
	hint := fs.Bool("hint", false, "print a hint for the provided board/string")
	puzzleS := fs.String("string", "", "solve: 81-char puzzle string (0 or . for empty)")
	puzzleF := fs.String("file", "", "solve: path to file containing 81-char puzzle string")
	asJSON := fs.Bool("json", false, "print output as JSON")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, versionString())
		return 0
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")

	if *puzzleS != "" || *puzzleF != "" {
		s := *puzzleS
		if *puzzleF != "" {
			b, err := os.ReadFile(*puzzleF)
			if err != nil {
				fmt.Fprintln(stderr, "error:", err)
				return 1
			}
			s = strings.TrimSpace(string(b))
		}
		board, err := sudoku.FromString(strings.TrimSpace(s))
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if *hint {
			r, c, v, ok := sudoku.Hint(board)
			if !ok {
				fmt.Fprintln(stderr, "error:", "no hint available")
				return 1
			}
			if *asJSON {
				_ = enc.Encode(map[string]int{"row": r, "col": c, "val": v})
			} else {
				fmt.Fprintf(stdout, "Hint: row %d, col %d = %d\n", r+1, c+1, v)
			}
			return 0
		}
		solved, ok := sudoku.Solve(board)
		if !ok {
			fmt.Fprintln(stderr, "error:", "unsolvable puzzle")
			return 1
		}
		if *asJSON {
			_ = enc.Encode(map[string]any{"solution": solved})
			return 0
		}
		fmt.Fprintln(stdout, "Solution:")
		printBoardTo(stdout, solved)
		return 0
	}

	var d sudoku.Difficulty
	switch strings.ToLower(*diff) {
	case string(sudoku.Easy):
		d = sudoku.Easy
	case string(sudoku.Medium), "":
		d = sudoku.Medium
	case string(sudoku.Hard):
		d = sudoku.Hard
	default:
		fmt.Fprintln(stderr, "error:", fmt.Errorf("invalid difficulty: %s", *diff))
		return 2
	}

	var br, bc int
	if _, err := fmt.Sscanf(*box, "%dx%d", &br, &bc); err != nil || br <= 0 || bc <= 0 || br*bc != *size {
		fmt.Fprintln(stderr, "error:", errors.New("invalid box dims; ensure size == R*C"))
		return 2
	}
	if *size == 9 && br == 3 && bc == 3 {
		puz, err := sudoku.Generate(d, *attempts)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if *asJSON {
			out := map[string]any{"puzzle": puz}
			if *showSol {
				if sol, ok := sudoku.Solve(puz); ok {
					out["solution"] = sol
				}
			}
			_ = enc.Encode(out)
			return 0
		}
		fmt.Fprintf(stdout, "Generated (%s):\n", d)
		printBoardTo(stdout, puz)
		if *showSol {
			if sol, ok := sudoku.Solve(puz); ok {
				fmt.Fprintln(stdout, "\nSolution:")
				printBoardTo(stdout, sol)
			}
		}
		return 0
	}
	g, err := sudoku.NewGrid(*size, br, bc)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	gpuz, err := g.Generate(d, *attempts)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if *asJSON {
		out := struct {
			Size  int    `json:"size"`
			BoxR  int    `json:"boxR"`
			BoxC  int    `json:"boxC"`
			Board string `json:"board"`
		}{gpuz.Size, gpuz.BoxRows, gpuz.BoxCols, gpuz.String()}
		_ = enc.Encode(out)
		return 0
	}
	fmt.Fprintf(stdout, "%dx%d (%dx%d boxes)\n", gpuz.Size, gpuz.Size, gpuz.BoxRows, gpuz.BoxCols)
	for r := 0; r < gpuz.Size; r++ {
		for c := 0; c < gpuz.Size; c++ {
			v := gpuz.Cells[r][c]
			if v == 0 {
				fmt.Fprint(stdout, ".")
			} else {
				fmt.Fprint(stdout, v)
			}
			if c < gpuz.Size-1 {
				fmt.Fprint(stdout, " ")
			}
		}
		fmt.Fprintln(stdout)
	}
	return 0
}

func printBoard(b sudoku.Board) {
	line := "+-------+-------+-------+"
	fmt.Println(line)
	for r := 0; r < 9; r++ {
		fmt.Print("|")
		for c := 0; c < 9; c++ {
			v := b[r][c]
			ch := '.'
			if v != 0 {
				ch = rune('0' + v)
			}
			sep := " "
			if (c+1)%3 == 0 {
				sep = " |"
			}
			fmt.Printf(" %c%s", ch, sep)
		}
		fmt.Println()
		if (r+1)%3 == 0 {
			fmt.Println(line)
		}
	}
}

func printBoardTo(w io.Writer, b sudoku.Board) {
	line := "+-------+-------+-------+"
	fmt.Fprintln(w, line)
	for r := 0; r < 9; r++ {
		fmt.Fprint(w, "|")
		for c := 0; c < 9; c++ {
			v := b[r][c]
			ch := '.'
			if v != 0 {
				ch = rune('0' + v)
			}
			sep := " "
			if (c+1)%3 == 0 {
				sep = " |"
			}
			fmt.Fprintf(w, " %c%s", ch, sep)
		}
		fmt.Fprintln(w)
		if (r+1)%3 == 0 {
			fmt.Fprintln(w, line)
		}
	}
}

func readAll(r io.Reader) string {
	sc := bufio.NewScanner(r)
	var sb strings.Builder
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		sb.WriteString(s)
	}
	return sb.String()
}

func check(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
