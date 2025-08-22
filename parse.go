package sudoku

import (
	"errors"
)

// FromString parses an 81-char string into a Board. Digits 1-9 are values, 0 or '.' are empty.
func FromString(s string) (Board, error) {
	var b Board
	if len(s) != 81 {
		return b, errors.New("input must be 81 characters")
	}
	for i := 0; i < 81; i++ {
		ch := s[i]
		r := i / 9
		c := i % 9
		switch ch {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			b[r][c] = int(ch - '0')
		case '0', '.':
			b[r][c] = 0
		default:
			return Board{}, errors.New("invalid character in board")
		}
	}
	if err := Validate(b); err != nil {
		return Board{}, err
	}
	return b, nil
}

// String returns 81-char representation of the board, '0' for empty.
func (b Board) String() string {
	buf := make([]byte, 0, 81)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			v := b[r][c]
			if v == 0 {
				buf = append(buf, '0')
			} else {
				buf = append(buf, byte('0'+v))
			}
		}
	}
	return string(buf)
}
