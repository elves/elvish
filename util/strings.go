package util

import (
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
