package strutil

import (
	"unicode"
	"unicode/utf8"
)

// Title returns s with the first codepoint changed to title case.
func Title(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size == 0 {
		return s
	}
	return string(unicode.ToTitle(r)) + s[size:]
}
