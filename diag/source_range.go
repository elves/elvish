package diag

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/elves/elvish/util"
)

// Context is a range of text in a source code. It is typically used for
// errors that can be associated with a part of the source code, like parse
// errors and a traceback entry.
type Context struct {
	Name   string
	Source string
	Begin  int
	End    int

	savedPPrintInfo *rangePPrintInfo
}

// NewContext creates a new Context.
func NewContext(name, source string, begin, end int) *Context {
	return &Context{name, source, begin, end, nil}
}

// Range returns the range of the Context.
func (sr *Context) Range() Ranging {
	return Ranging{sr.Begin, sr.End}
}

// Information about the source range that are needed for pretty-printing.
type rangePPrintInfo struct {
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

func (sr *Context) pprintInfo() *rangePPrintInfo {
	if sr.savedPPrintInfo != nil {
		return sr.savedPPrintInfo
	}

	before := sr.Source[:sr.Begin]
	culprit := sr.Source[sr.Begin:sr.End]
	after := sr.Source[sr.End:]

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

	sr.savedPPrintInfo = &rangePPrintInfo{head, culprit, tail, beginLine, endLine}
	return sr.savedPPrintInfo
}

// PPrint pretty-prints a SourceContext.
func (sr *Context) PPrint(sourceIndent string) string {
	if err := sr.checkPosition(); err != nil {
		return err.Error()
	}
	return (sr.Name + ", " + sr.lineRange() +
		"\n" + sourceIndent + sr.relevantSource(sourceIndent))
}

// PPrintCompact pretty-prints a SourceContext, with no line break between the
// source position range description and relevant source excerpt.
func (sr *Context) PPrintCompact(sourceIndent string) string {
	if err := sr.checkPosition(); err != nil {
		return err.Error()
	}
	desc := sr.Name + ", " + sr.lineRange() + " "
	// Extra indent so that following lines line up with the first line.
	descIndent := strings.Repeat(" ", util.Wcswidth(desc))
	return desc + sr.relevantSource(sourceIndent+descIndent)
}

func (sr *Context) checkPosition() error {
	if sr.Begin == -1 {
		return fmt.Errorf("%s, unknown position", sr.Name)
	} else if sr.Begin < 0 || sr.End > len(sr.Source) || sr.Begin > sr.End {
		return fmt.Errorf("%s, invalid position %d-%d", sr.Name, sr.Begin, sr.End)
	}
	return nil
}

func (sr *Context) lineRange() string {
	info := sr.pprintInfo()

	if info.BeginLine == info.EndLine {
		return fmt.Sprintf("line %d:", info.BeginLine)
	}
	return fmt.Sprintf("line %d-%d:", info.BeginLine, info.EndLine)
}

func (sr *Context) relevantSource(sourceIndent string) string {
	info := sr.pprintInfo()

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
