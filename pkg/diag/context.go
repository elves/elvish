package diag

import (
	"fmt"
	"strings"
)

// Context stores information derived from a range in some text. It is used for
// errors that point to a part of the source code, including parse errors,
// compilation errors and a single traceback entry in an exception.
//
// Context values should only be constructed using [NewContext].
type Context struct {
	Name string
	Ranging
	// 1-based line and column numbers of the start position.
	StartLine, StartCol int
	// 1-based line and column numbers of the end position, inclusive. Note that
	// if the range is zero-width, EndCol will be StartCol - 1.
	EndLine, EndCol int
	// The relevant text, text before its the first line and the text after its
	// last line.
	Body, Head, Tail string
}

// NewContext creates a new Context.
func NewContext(name, source string, r Ranger) *Context {
	rg := r.Range()
	d := getContextDetails(source, rg)
	return &Context{name, rg,
		d.startLine, d.startCol, d.endLine, d.endCol, d.body, d.head, d.tail}
}

// Show shows the context.
//
// If the body has only one line, it returns one line like:
//
//	foo.elv:12:7-11: lorem ipsum
//
// If the body has multiple lines, it shows the body in an indented block:
//
//	foo.elv:12:1-13:5
//	  lorem
//	  ipsum
//
// The body is underlined.
func (c *Context) Show(indent string) string {
	rangeDesc := c.describeRange()
	if c.StartLine == c.EndLine {
		// Body has only one line, show it on the same line:
		//
		return fmt.Sprintf("%s: %s",
			rangeDesc, showContextText(indent, c.Head, c.Body, c.Tail))
	}
	indent += "  "
	return fmt.Sprintf("%s:\n%s%s",
		rangeDesc, indent, showContextText(indent, c.Head, c.Body, c.Tail))
}

func (c *Context) describeRange() string {
	if c.StartLine == c.EndLine {
		if c.EndCol < c.StartCol {
			// Since EndCol is inclusive, zero-width ranges result in EndCol =
			// StartCol - 1.
			return fmt.Sprintf("%s:%d:%d", c.Name, c.StartLine, c.StartCol)
		}
		return fmt.Sprintf("%s:%d:%d-%d",
			c.Name, c.StartLine, c.StartCol, c.EndCol)
	}
	return fmt.Sprintf("%s:%d:%d-%d:%d",
		c.Name, c.StartLine, c.StartCol, c.EndLine, c.EndCol)
}

// Variables controlling the style used in [*Context.Show]. Can be overridden in
// tests.
var (
	ContextBodyStartMarker = "\033[1;4m"
	ContextBodyEndMarker   = "\033[m"
)

func showContextText(indent, head, body, tail string) string {
	var sb strings.Builder
	sb.WriteString(head)

	for i, line := range strings.Split(body, "\n") {
		if i > 0 {
			sb.WriteByte('\n')
			sb.WriteString(indent)
		}
		sb.WriteString(ContextBodyStartMarker)
		sb.WriteString(line)
		sb.WriteString(ContextBodyEndMarker)
	}

	sb.WriteString(tail)
	return sb.String()
}

// Information about the lines that contain the culprit.
type contextDetails struct {
	startLine, startCol int
	endLine, endCol     int
	body, head, tail    string
}

func getContextDetails(source string, r Ranging) contextDetails {
	before := source[:r.From]
	body := source[r.From:r.To]
	after := source[r.To:]

	head := lastLine(before)

	// If the body ends with a newline, stripe it, and leave the tail empty.
	// Otherwise, don't process the body and calculate the tail.
	var tail string
	if strings.HasSuffix(body, "\n") {
		body = body[:len(body)-1]
	} else {
		tail = firstLine(after)
	}

	startLine := strings.Count(before, "\n") + 1
	startCol := 1 + len(head)
	endLine := startLine + strings.Count(body, "\n")
	var endCol int
	if startLine == endLine {
		endCol = startCol + len(body) - 1
	} else {
		endCol = len(lastLine(body))
	}

	return contextDetails{startLine, startCol, endLine, endCol, body, head, tail}
}

func firstLine(s string) string {
	i := strings.IndexByte(s, '\n')
	if i == -1 {
		return s
	}
	return s[:i]
}

func lastLine(s string) string {
	// When s does not contain '\n', LastIndexByte returns -1, which happens to
	// be what we want.
	return s[strings.LastIndexByte(s, '\n')+1:]
}
