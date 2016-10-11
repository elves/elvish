package util

import (
	"fmt"
	"io"
	"strings"
)

type TracebackEntry struct {
	Name   string
	Source string
	Begin  int
	End    int
}

var CulpritStyle = "1;4"

func (te *TracebackEntry) Pprint(w io.Writer, sourceIndent string) {
	if te.Begin == -1 {
		fmt.Fprintf(w, "%s, unknown position", te.Name)
		return
	} else if te.Begin < 0 || te.End > len(te.Source) || te.Begin > te.End {
		fmt.Fprintf(w, "%s, invalid position", te.Name)
		return
	}

	before, culprit, after := bca(te.Source, te.Begin, te.End)
	// Find the part of "before" that is on the same line as the culprit.
	lineBefore := lastLine(before)
	// Find on which line the culprit begins.
	beginLine := strings.Count(before, "\n") + 1
	// Find on which column the culprit begins.
	beginCol := countRunes(lineBefore) + 1

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
	var endCol int
	if endLine == beginLine {
		endCol = beginCol + countRunes(culprit) - 1
	} else {
		endCol = countRunes(lastLine(culprit))
	}

	if beginLine == endLine {
		if endCol <= beginCol {
			fmt.Fprintf(w, "%s, line %d, col %d:\n", te.Name, beginLine, beginCol)
		} else {
			fmt.Fprintf(w, "%s, line %d, col %d-%d:\n", te.Name, beginLine, beginCol, endCol)
		}
	} else {
		fmt.Fprintf(w, "%s, line %d col %d - line %d col %d:\n", te.Name, beginLine, beginCol, endLine, endCol)
	}

	fmt.Fprintf(w, "%s%s", sourceIndent, lineBefore)

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

func countRunes(s string) int {
	return strings.Count(s, "") - 1
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
