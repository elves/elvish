// errutil provides an exception-like mechanism and the ContextualError type.
package errutil

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/elves/elvish/util"
)

type ContextualError struct {
	srcname string
	title   string
	line    string
	lineno  int
	colno   int
	msg     string
}

func NewContextualError(srcname, title, text string, pos int, format string, args ...interface{}) *ContextualError {
	lineno, colno, line := util.FindContext(text, pos)
	return &ContextualError{srcname, title, line, lineno, colno, fmt.Sprintf(format, args...)}
}

func (e *ContextualError) Error() string {
	return fmt.Sprintf("%s:%d:%d %s:%s", e.srcname, e.lineno, e.colno, e.title, e.msg)
}

func (e *ContextualError) Pprint() string {
	buf := new(bytes.Buffer)
	// Position info
	fmt.Fprintf(buf, "\033[1m%s:%d:%d: ", e.srcname, e.lineno+1, e.colno+1)
	// "error:"
	fmt.Fprintf(buf, "\033[31m%s: ", e.title)
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
