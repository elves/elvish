package strutil

import (
	"strings"
)

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
