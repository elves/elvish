package util

import (
	"errors"
	"strings"
)

// ErrIndexOutOfRange is returned when out-of-range errors occur.
var ErrIndexOutOfRange = errors.New("substring out of range")

// FindContext takes a position in a text and finds its line number,
// corresponding line and column numbers. Line and column numbers are counted
// from 0. Used in diagnostic messages.
func FindContext(text string, pos int) (lineno, colno int, line string) {
	var i, linestart int
	var r rune
	for i, r = range text {
		if i == pos {
			break
		}
		if r == '\n' {
			lineno++
			linestart = i + 1
			colno = 0
		} else {
			colno++
		}
	}
	line = strings.SplitN(text[linestart:], "\n", 2)[0]
	return
}

// FindFirstEOL returns the index of the first '\n'. When there is no '\n', the
// length of s is returned.
func FindFirstEOL(s string) int {
	eol := strings.IndexRune(s, '\n')
	if eol == -1 {
		eol = len(s)
	}
	return eol
}

// FindLastSOL returns an index just after the last '\n'.
func FindLastSOL(s string) int {
	return strings.LastIndex(s, "\n") + 1
}

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

// MatchSubseq returns whether pattern is a subsequence of s.
func MatchSubseq(s, pattern string) bool {
	for _, p := range pattern {
		i := strings.IndexRune(s, p)
		if i == -1 {
			return false
		}
		s = s[i+len(string(p)):]
	}
	return true
}
