package diag

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/wcwidth"
)

// Context is a range of text in a source code. It is typically used for
// errors that can be associated with a part of the source code, like parse
// errors and a traceback entry.
//
// Context values should only be constructed using [NewContext].
type Context struct {
	Name   string
	Source string
	Ranging

	culprit culprit
}

// NewContext creates a new Context.
func NewContext(name, source string, r Ranger) *Context {
	rg := r.Range()
	return &Context{name, source, rg, makeCulprit(source, rg)}
}

// Show shows a SourceContext.
func (c *Context) Show(indent string) string {
	return fmt.Sprintf("%s:%s:\n%s%s",
		c.Name, c.culprit.describeStart(), indent+"  ", c.culprit.Show(indent+"  "))
}

// ShowCompact shows a Context, with no line break between the culprit range
// description and relevant source excerpt.
func (c *Context) ShowCompact(indent string) string {
	desc := fmt.Sprintf("%s:%s: ", c.Name, c.culprit.describeStart())
	// Extra indent so that following lines line up with the first line.
	descIndent := strings.Repeat(" ", wcwidth.Of(desc))
	return desc + c.culprit.Show(indent+descIndent)
}

// Information about the lines that contain the culprit.
type culprit struct {
	// The actual culprit text.
	Body string
	// Text before Body on its first line.
	Head string
	// Text after Body on its last line.
	Tail string
	// 1-based line and column numbers of the start position.
	StartLine, StartCol int
}

func makeCulprit(source string, r Ranging) culprit {
	before := source[:r.From]
	body := source[r.From:r.To]
	after := source[r.To:]

	head := lastLine(before)
	fromLine := strings.Count(before, "\n") + 1
	fromCol := 1 + wcwidth.Of(head)

	// If the culprit ends with a newline, stripe it, and tail is empty.
	// Otherwise, tail is nonempty.
	var tail string
	if strings.HasSuffix(body, "\n") {
		body = body[:len(body)-1]
	} else {
		tail = firstLine(after)
	}

	return culprit{body, head, tail, fromLine, fromCol}
}

// Variables controlling the style of the culprit.
var (
	culpritStart       = "\033[1;4m"
	culpritEnd         = "\033[m"
	culpritPlaceHolder = "^"
)

func (cl *culprit) describeStart() string {
	return fmt.Sprintf("%d:%d", cl.StartLine, cl.StartCol)
}

func (cl *culprit) Show(indent string) string {
	var sb strings.Builder
	sb.WriteString(cl.Head)

	body := cl.Body
	if body == "" {
		body = culpritPlaceHolder
	}

	for i, line := range strings.Split(body, "\n") {
		if i > 0 {
			sb.WriteByte('\n')
			sb.WriteString(indent)
		}
		sb.WriteString(culpritStart)
		sb.WriteString(line)
		sb.WriteString(culpritEnd)
	}

	sb.WriteString(cl.Tail)
	return sb.String()
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
