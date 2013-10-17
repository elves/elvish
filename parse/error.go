package parse

import (
	"fmt"
	"bytes"
	"strings"
)

// Given a position in a text, find its line number, corresponding line and
// column numbers. Line and column numbers are counted from 0.
// Used in diagnostic messages.
func findContext(text string, pos int) (lineno, colno int, line string) {
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

type Error struct {
	name string
	lineno int
	colno int
	line string
	msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d %s", e.name, e.lineno, e.colno, e.msg)
}

func (e *Error) Pprint() string {
	buf := new(bytes.Buffer)
	// Position info
	fmt.Fprintf(buf, "\033[1m%s:%d:%d: ", e.name, e.lineno+1, e.colno+1)
	// "error:"
	fmt.Fprintf(buf, "\033[31merror: ")
	// Message
	fmt.Fprintf(buf, "\033[m\033[1m%s\033[m\n", e.msg)
	// Context: line
	// TODO Handle long lines
	fmt.Fprintf(buf, "%s\n", e.line)
	// Context: arrow
	// TODO Handle multi-width characters
	fmt.Fprintf(buf, "%s\033[32;1m^\033[m\n", strings.Repeat(" ", e.colno))
	return buf.String()
}
