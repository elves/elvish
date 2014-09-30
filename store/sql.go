package store

import "strings"

var globEscaper = strings.NewReplacer("\\", "\\\\", "?", "\\?", "*", "\\*")

// EscapeGlob escapes s to be suitable as an argument to SQLite's GLOB
// operator.
func EscapeGlob(s string) string {
	return globEscaper.Replace(s)
}
