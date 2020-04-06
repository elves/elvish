package diag

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/elves/elvish/pkg/wcwidth"
)

// Context is a range of text in a source code. It is typically used for
// errors that can be associated with a part of the source code, like parse
// errors and a traceback entry.
type Context struct {
	Name   string
	Source string
	Ranging

	savedShowInfo *rangeShowInfo
}

// NewContext creates a new Context.
func NewContext(name, source string, r Ranger) *Context {
	return &Context{name, source, r.Range(), nil}
}

// Information about the source range that are needed for showing.
type rangeShowInfo struct {
	// Head is the piece of text immediately before Culprit, extending to, but
	// not including the closest line boundary. If Culprit already starts after
	// a line boundary, Head is an empty string.
	Head string
	// Culprit is Source[Begin:End], with any trailing newlines stripped.
	Culprit string
	// Tail is the piece of text immediately after Culprit, extending to, but
	// not including the closet line boundary. If Culprit already ends before a
	// line boundary, Tail is an empty string.
	Tail string
	// BeginLine is the (1-based) line number that the first character of Culprit is on.
	BeginLine int
	// EndLine is the (1-based) line number that the last character of Culprit is on.
	EndLine int
}

// Variables controlling the style of the culprit.
var (
	culpritLineBegin   = "\033[1;4m"
	culpritLineEnd     = "\033[m"
	culpritPlaceHolder = "^"
)

func (c *Context) showInfo() *rangeShowInfo {
	if c.savedShowInfo != nil {
		return c.savedShowInfo
	}

	before := c.Source[:c.From]
	culprit := c.Source[c.From:c.To]
	after := c.Source[c.To:]

	head := lastLine(before)
	beginLine := strings.Count(before, "\n") + 1

	// If the culprit ends with a newline, stripe it. Otherwise, tail is nonempty.
	var tail string
	if strings.HasSuffix(culprit, "\n") {
		culprit = culprit[:len(culprit)-1]
	} else {
		tail = firstLine(after)
	}

	endLine := beginLine + strings.Count(culprit, "\n")

	c.savedShowInfo = &rangeShowInfo{head, culprit, tail, beginLine, endLine}
	return c.savedShowInfo
}

// Show shows a SourceContext.
func (c *Context) Show(sourceIndent string) string {
	if err := c.checkPosition(); err != nil {
		return err.Error()
	}
	return (c.Name + ", " + c.lineRange() +
		"\n" + sourceIndent + c.relevantSource(sourceIndent))
}

// ShowCompact shows a SourceContext, with no line break between the
// source position range description and relevant source excerpt.
func (c *Context) ShowCompact(sourceIndent string) string {
	if err := c.checkPosition(); err != nil {
		return err.Error()
	}
	desc := c.Name + ", " + c.lineRange() + " "
	// Extra indent so that following lines line up with the first line.
	descIndent := strings.Repeat(" ", wcwidth.Of(desc))
	return desc + c.relevantSource(sourceIndent+descIndent)
}

func (c *Context) checkPosition() error {
	if c.From == -1 {
		return fmt.Errorf("%s, unknown position", c.Name)
	} else if c.From < 0 || c.To > len(c.Source) || c.From > c.To {
		return fmt.Errorf("%s, invalid position %d-%d", c.Name, c.From, c.To)
	}
	return nil
}

func (c *Context) lineRange() string {
	info := c.showInfo()

	if info.BeginLine == info.EndLine {
		return fmt.Sprintf("line %d:", info.BeginLine)
	}
	return fmt.Sprintf("line %d-%d:", info.BeginLine, info.EndLine)
}

func (c *Context) relevantSource(sourceIndent string) string {
	info := c.showInfo()

	var buf bytes.Buffer
	buf.WriteString(info.Head)

	culprit := info.Culprit
	if culprit == "" {
		culprit = culpritPlaceHolder
	}

	for i, line := range strings.Split(culprit, "\n") {
		if i > 0 {
			buf.WriteByte('\n')
			buf.WriteString(sourceIndent)
		}
		buf.WriteString(culpritLineBegin)
		buf.WriteString(line)
		buf.WriteString(culpritLineEnd)
	}

	buf.WriteString(info.Tail)
	return buf.String()
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
