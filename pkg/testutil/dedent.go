package testutil

// This code is a fork of https://github.com/lithammer/dedent, but with a fix
// for https://github.com/lithammer/dedent/issues/20 so that we can use raw
// strings in a natural manner by placing the first indented line on a new line
// following the opening backtick.

import (
	"regexp"
	"strings"
)

var (
	whitespaceOnly    = regexp.MustCompile("(?m)^[ \t]+$")
	leadingWhitespace = regexp.MustCompile("(?m)(^[ \t]*)(?:[^ \t\n])")
)

// Dedent removes any common leading whitespace from every line in text. An
// initial newline is removed.
//
// This can be used to make multiline (usually raw) strings to line up with the
// left edge of the display, while still presenting them in the source code in
// indented form.
func Dedent(text string) string {
	var margin string

	if text[0] == '\n' {
		text = whitespaceOnly.ReplaceAllString(text[1:], "")
	} else {
		text = whitespaceOnly.ReplaceAllString(text, "")
	}
	indents := leadingWhitespace.FindAllStringSubmatch(text, -1)

	// Look for the longest leading string of spaces and tabs common to all
	// lines.
	for i, indent := range indents {
		if i == 0 {
			margin = indent[1]
		} else if strings.HasPrefix(indent[1], margin) {
			// Current line more deeply indented than previous winner:
			// no change (previous winner is still on top).
			continue
		} else if strings.HasPrefix(margin, indent[1]) {
			// Current line consistent with and no deeper than previous winner:
			// it's the new winner.
			margin = indent[1]
		} else {
			// Current line and previous winner have no common whitespace:
			// there is no margin.
			margin = ""
			break
		}
	}

	if margin != "" {
		text = regexp.MustCompile("(?m)^"+margin).ReplaceAllString(text, "")
	}
	return text
}
