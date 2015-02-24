package util

import (
	"errors"
	"strings"
)

// FindContext takes a position in a text and finds its line number,
// corresponding line and column numbers. Line and column numbers are counted
// from 0. Used in diagnostic messages.
func FindContext(text string, pos int) (lineno, colno int, line string) {
	var p int
	for _, r := range text {
		if p == pos {
			break
		}
		if r == '\n' {
			lineno++
			colno = 0
		} else {
			colno++
		}
		p++
	}
	line = strings.SplitN(text[p-colno:], "\n", 2)[0]
	return
}

func FindFirstEOL(s string) int {
	eol := strings.IndexRune(s, '\n')
	if eol == -1 {
		eol = len(s)
	}
	return eol
}

func FindLastSOL(s string) int {
	return strings.LastIndex(s, "\n") + 1
}

var (
	ErrIndexOutOfRange = errors.New("substring out of range")
)

// SubstringByRune returns the range of the i-th rune (inclusive) through the
// j-th rune (exclusive) in s.
func SubstringByRune(s string, low, high int) (string, error) {
	if low > high || low < 0 || high < 0 {
		return "", ErrIndexOutOfRange
	}
	var bLow, bHigh, j int
	for i := range s {
		if j == low {
			bLow = i
		}
		if j == high {
			bHigh = i
		}
		j++
	}
	if j < high {
		return "", ErrIndexOutOfRange
	}
	if low == high {
		return "", nil
	}
	if j == high {
		bHigh = len(s)
	}
	return s[bLow:bHigh], nil
}

// NthRune returns the n-th rune of s.
func NthRune(s string, n int) (rune, error) {
	if n < 0 {
		return 0, ErrIndexOutOfRange
	}
	var j int
	for _, r := range s {
		if j == n {
			return r, nil
		}
		j++
	}
	return 0, ErrIndexOutOfRange
}
