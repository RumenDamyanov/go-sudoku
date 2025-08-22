package sudoku

import "testing"

func TestFromStringErrors(t *testing.T) {
	// wrong length
	if _, err := FromString("123"); err == nil {
		t.Fatalf("expected length error")
	}
	// invalid character
	bad := make([]byte, 81)
	for i := range bad {
		bad[i] = '0'
	}
	bad[10] = 'x'
	if _, err := FromString(string(bad)); err == nil {
		t.Fatalf("expected invalid character error")
	}
	// duplicate in a row should fail via Validate
	dup := "11" + makeStr('0', 79)
	if _, err := FromString(dup); err == nil {
		t.Fatalf("expected validate error for duplicates")
	}
}

func TestBoardStringRoundtrip(t *testing.T) {
	in := "53..7....6..195....98....6.8...6...34..8.3..17...2...6.6....28....419..5....8..79"
	b, err := FromString(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out := b.String()
	if len(out) != 81 {
		t.Fatalf("got len=%d", len(out))
	}
}

func makeStr(ch byte, n int) string {
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = ch
	}
	return string(buf)
}
