package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.rumenx.com/sudoku"
)


var (
	// override with -ldflags "-X main.version=... -X main.commit=... -X main.date=..."
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": version, "commit": commit, "date": date})
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { // alias
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": version, "commit": commit, "date": date})
	})
	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/solve", handleSolve)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	s := &http.Server{
		Addr:              addr,
		Handler:           logRequest(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("listening on %s", addr)
	log.Fatal(s.ListenAndServe())
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errMsg("method not allowed"))
		return
	}
	var req struct {
		Difficulty      string `json:"difficulty"`
		IncludeSolution bool   `json:"includeSolution"`
		Size            int    `json:"size"`
		Box             string `json:"box"`      // e.g. 3x3, 2x3
		Attempts        int    `json:"attempts"` // generation attempts
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid json"))
		return
	}
	var d sudoku.Difficulty
	switch req.Difficulty {
	case string(sudoku.Easy), "":
		d = sudoku.Easy
	case string(sudoku.Medium):
		d = sudoku.Medium
	case string(sudoku.Hard):
		d = sudoku.Hard
	default:
		writeJSON(w, http.StatusBadRequest, errMsg("invalid difficulty"))
		return
	}
	if req.Attempts < 1 {
		req.Attempts = 3
	}
	if req.Size == 0 && req.Box == "" { // classic 9x9 shortcut
		puz, err := sudoku.Generate(d, req.Attempts)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errMsg("generation failed"))
			return
		}
		res := map[string]any{"puzzle": puz}
		if req.IncludeSolution {
			if sol, ok := sudoku.Solve(puz); ok {
				res["solution"] = sol
			}
		}
		writeJSON(w, http.StatusOK, res)
		return
	}
	// variable size path
	if req.Size <= 0 || req.Box == "" {
		writeJSON(w, http.StatusBadRequest, errMsg("size and box required for variable grid"))
		return
	}
	if req.Size > sudoku.MaxGridSize {
		writeJSON(w, http.StatusBadRequest, errMsg(fmt.Sprintf(
			"grid size %d exceeds maximum allowed (%d)", req.Size, sudoku.MaxGridSize)))
		return
	}
	var br, bc int
	if _, err := fmt.Sscanf(req.Box, "%dx%d", &br, &bc); err != nil || br*bc != req.Size {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid box dims"))
		return
	}
	g, err := sudoku.NewGrid(req.Size, br, bc)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid grid params"))
		return
	}
	gpuz, err := g.Generate(d, req.Attempts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMsg("generation failed"))
		return
	}
	res := map[string]any{
		"size":   gpuz.Size,
		"boxR":   gpuz.BoxRows,
		"boxC":   gpuz.BoxCols,
		"puzzle": gpuz.Cells,
	}
	writeJSON(w, http.StatusOK, res)
}

func handleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errMsg("method not allowed"))
		return
	}
	var req struct {
		Puzzle *sudoku.Board `json:"puzzle"`
		String string        `json:"string"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errMsg("invalid json"))
		return
	}
	var b sudoku.Board
	var err error
	if req.Puzzle != nil {
		b = *req.Puzzle
		if err = sudoku.Validate(b); err != nil {
			writeJSON(w, http.StatusBadRequest, errMsg("invalid puzzle"))
			return
		}
	} else if req.String != "" {
		if b, err = sudoku.FromString(req.String); err != nil {
			writeJSON(w, http.StatusBadRequest, errMsg("invalid puzzle string"))
			return
		}
	} else {
		writeJSON(w, http.StatusBadRequest, errMsg("missing puzzle"))
		return
	}
	if sol, ok := sudoku.Solve(b); ok {
		writeJSON(w, http.StatusOK, map[string]any{"solution": sol})
		return
	}
	writeJSON(w, http.StatusUnprocessableEntity, errMsg("unsolvable"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errMsg(msg string) map[string]string { return map[string]string{"error": msg} }

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()
		// propagate context to handlers (already using r directly)
		next.ServeHTTP(w, r.WithContext(ctx))
		dur := time.Since(start)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, dur)
	})
}
