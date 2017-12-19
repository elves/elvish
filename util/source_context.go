package util

import (
	"fmt"
	"io"
	"strings"
)

// SourceContext is a range of text in a source code. It can point to another
// SourceContext, thus forming a linked list. It is used for tracebacks.
type SourceContext struct {
	Name   string
	Source string
	Begin  int
	End    int
	Next   *SourceContext

	savedPprintInfo *contextPprintInfo
}

func NewSourceContext(name, source string, begin, end int, next *SourceContext) *SourceContext {
	return &SourceContext{name, source, begin, end, next, nil}
}

// contextPprintInfo is information about the source context that are friendly to
// human, used when pretty-printing.
type contextPprintInfo struct {
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
	CulpritStyle       = "1;4"
	CulpritPlaceHolder = "^"
)

func (sc *SourceContext) pprintInfo() *contextPprintInfo {
	if sc.savedPprintInfo != nil {
		return sc.savedPprintInfo
	}

	before := sc.Source[:sc.Begin]
	culprit := sc.Source[sc.Begin:sc.End]
	after := sc.Source[sc.End:]

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

	sc.savedPprintInfo = &contextPprintInfo{head, culprit, tail, beginLine, endLine}
	return sc.savedPprintInfo
}

// Pprint pretty-prints a SourceContext.
func (sc *SourceContext) Pprint(w io.Writer, sourceIndent string) {
	if sc.complainBadPosition(w) {
		return
	}
	sc.printPosDescription(w)
	fmt.Fprintf(w, "\n%s%s", sourceIndent, "")
	sc.pprintRelevantSource(w, sourceIndent)
}

func (sc *SourceContext) complainBadPosition(w io.Writer) bool {
	if sc.Begin == -1 {
		fmt.Fprintf(w, "%s, unknown position", sc.Name)
		return true
	} else if sc.Begin < 0 || sc.End > len(sc.Source) || sc.Begin > sc.End {
		fmt.Fprintf(w, "%s, invalid position %d-%d", sc.Name, sc.Begin, sc.End)
		return true
	}
	return false
}

func (sc *SourceContext) printPosDescription(w io.Writer) {
	info := sc.pprintInfo()

	if info.BeginLine == info.EndLine {
		fmt.Fprintf(w, "%s, line %d:", sc.Name, info.BeginLine)
	} else {
		fmt.Fprintf(w, "%s, line %d-%d:", sc.Name, info.BeginLine, info.EndLine)
	}
}

func (sc *SourceContext) pprintRelevantSource(w io.Writer, sourceIndent string) {
	info := sc.pprintInfo()

	fmt.Fprint(w, info.Head)

	culprit := info.Culprit
	if culprit == "" {
		culprit = CulpritPlaceHolder
	}

	for i, line := range strings.Split(culprit, "\n") {
		if i > 0 {
			fmt.Fprintf(w, "\n%s", sourceIndent)
		}
		fmt.Fprintf(w, "\033[%sm%s\033[m", CulpritStyle, line)
	}

	fmt.Fprint(w, info.Tail)
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
