package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMuxForTest() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/solve", handleSolve)
	return mux
}

func TestHealthz(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestGenerateAPI(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	body, _ := json.Marshal(map[string]any{"difficulty": "medium", "includeSolution": true})
	resp, err := http.Post(ts.URL+"/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestGenerateAPI_WithSolution(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	body, _ := json.Marshal(map[string]any{"difficulty": "easy", "includeSolution": true})
	resp, err := http.Post(ts.URL+"/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestGenerateAPI_Errors(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	// wrong method
	resp, err := http.Get(ts.URL + "/generate")
	if err != nil {
		t.Fatalf("get generate: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	// bad difficulty
	body, _ := json.Marshal(map[string]any{"difficulty": "impossible"})
	resp, err = http.Post(ts.URL+"/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post generate: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSolveAPI(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	// known easy puzzle string
	s := "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	body, _ := json.Marshal(map[string]any{"string": s})
	resp, err := http.Post(ts.URL+"/solve", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("solve: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestSolveAPI_Errors(t *testing.T) {
	ts := httptest.NewServer(newMuxForTest())
	t.Cleanup(ts.Close)
	// method not allowed
	resp, err := http.Get(ts.URL + "/solve")
	if err != nil {
		t.Fatalf("get solve: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	// invalid json
	resp, err = http.Post(ts.URL+"/solve", "application/json", bytes.NewBufferString("{"))
	if err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	// missing puzzle
	resp, err = http.Post(ts.URL+"/solve", "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatalf("missing: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	// invalid string
	resp, err = http.Post(ts.URL+"/solve", "application/json", bytes.NewBufferString(`{"string":"xxx"}`))
	if err != nil {
		t.Fatalf("invalid string: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	// unsolvable: valid-shaped board with contradictions
	resp, err = http.Post(ts.URL+"/solve", "application/json", bytes.NewBufferString(`{"puzzle":[[1,1],[],[],[],[],[],[],[],[]]}`))
	if err != nil {
		t.Fatalf("unsolvable: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 400 or 422, got %d", resp.StatusCode)
	}
}
