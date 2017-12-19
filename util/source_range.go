package util

import (
	"fmt"
	"io"
	"strings"
)

// SourceRange is a range of text in a source code. It can point to another
// SourceRange, thus forming a linked list. It is used for tracebacks.
type SourceRange struct {
	Name   string
	Source string
	Begin  int
	End    int
	Next   *SourceRange

	savedPprintInfo *rangePprintInfo
}

// NewSourceRange creates a new SourceRange.
func NewSourceRange(name, source string, begin, end int, next *SourceRange) *SourceRange {
	return &SourceRange{name, source, begin, end, next, nil}
}

// rangePprintInfo is information about the source range that are needed for
// pretty-printing.
type rangePprintInfo struct {
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

func (sr *SourceRange) pprintInfo() *rangePprintInfo {
	if sr.savedPprintInfo != nil {
		return sr.savedPprintInfo
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

	sr.savedPprintInfo = &rangePprintInfo{head, culprit, tail, beginLine, endLine}
	return sr.savedPprintInfo
}

// Pprint pretty-prints a SourceContext.
func (sr *SourceRange) Pprint(w io.Writer, sourceIndent string) {
	if sr.complainBadPosition(w) {
		return
	}
	sr.printPosDescription(w)
	fmt.Fprintf(w, "\n%s%s", sourceIndent, "")
	sr.pprintRelevantSource(w, sourceIndent)
}

// PprintCompact pretty-prints a SourceContext, with no line break between the
// position description and relevant source excerpt.
func (sr *SourceRange) PprintCompact(w io.Writer, sourceIndent string) {
	if sr.complainBadPosition(w) {
		return
	}
	sr.printPosDescription(w)
	fmt.Fprint(w, ": ")
	// TODO sourceIndent += padding equal to what has been printed on this line
	sr.pprintRelevantSource(w, sourceIndent)
}

func (sr *SourceRange) complainBadPosition(w io.Writer) bool {
	if sr.Begin == -1 {
		fmt.Fprintf(w, "%s, unknown position", sr.Name)
		return true
	} else if sr.Begin < 0 || sr.End > len(sr.Source) || sr.Begin > sr.End {
		fmt.Fprintf(w, "%s, invalid position %d-%d", sr.Name, sr.Begin, sr.End)
		return true
	}
	return false
}

func (sr *SourceRange) printPosDescription(w io.Writer) {
	info := sr.pprintInfo()

	if info.BeginLine == info.EndLine {
		fmt.Fprintf(w, "%s, line %d:", sr.Name, info.BeginLine)
	} else {
		fmt.Fprintf(w, "%s, line %d-%d:", sr.Name, info.BeginLine, info.EndLine)
	}
}

func (sr *SourceRange) pprintRelevantSource(w io.Writer, sourceIndent string) {
	info := sr.pprintInfo()

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
