package util

import (
	"fmt"
	"io"
	"strings"
)

type SourceContext struct {
	Name   string
	Source string
	Begin  int
	End    int
	Next   *SourceContext
}

var CulpritStyle = "1;4"

func (sc *SourceContext) Pprint(w io.Writer, sourceIndent string) {
	sc.PprintOption(w, sourceIndent, true)
}

func (sc *SourceContext) PprintOption(w io.Writer, sourceIndent string, printContext bool) {
	if sc.Begin == -1 {
		fmt.Fprintf(w, "%s, unknown position", sc.Name)
		return
	} else if sc.Begin < 0 || sc.End > len(sc.Source) || sc.Begin > sc.End {
		fmt.Fprintf(w, "%s, invalid position %d-%d", sc.Name, sc.Begin, sc.End)
		return
	}

	before, culprit, after := bca(sc.Source, sc.Begin, sc.End)
	// Find the part of "before" that is on the same line as the culprit.
	lineBefore := lastLine(before)
	// Find on which line the culprit begins.
	beginLine := strings.Count(before, "\n") + 1

	// If the culprit ends with a newline, stripe it. Otherwise stick the part
	// of "after" that is on the same line of the last line of the culprit.
	var lineAfter string
	if strings.HasSuffix(culprit, "\n") {
		culprit = culprit[:len(culprit)-1]
	} else {
		lineAfter = firstLine(after)
	}

	// Find on which line and column the culprit ends.
	endLine := beginLine + strings.Count(culprit, "\n")

	if printContext {
		fmt.Fprintln(w, "%s, ", sc.Name)
	}

	// Save line %d expansion on a string, in case
	// its length is needed for indentation (printContext=false).
	var culpritLines string
	if beginLine == endLine {
		culpritLines = fmt.Sprintf("line %d:", beginLine)
	} else {
		culpritLines = fmt.Sprintf("line %d:%d", beginLine, endLine)
	}
	fmt.Fprintf(w, culpritLines)
	if printContext {
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "%s%s", sourceIndent, lineBefore)

	if !printContext {
		// If more lines come, they will be properly indented.
		sourceIndent += strings.Repeat(" ", len(culpritLines))
	}
	if culprit == "" {
		culprit = "^"
	}
	for i, line := range strings.Split(culprit, "\n") {
		if i > 0 {
			fmt.Fprintf(w, "\n%s", sourceIndent)
		}
		fmt.Fprintf(w, "\033[%sm%s\033[m", CulpritStyle, line)
	}

	fmt.Fprintf(w, "%s", lineAfter)
}

func bca(s string, a, b int) (string, string, string) {
	return s[:a], s[a:b], s[b:]
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
