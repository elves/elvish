package util

import (
	"fmt"
	"io"
	"strings"
)

type Traceback struct {
	Name   string
	Source string
	Begin  int
	End    int
	Next   *Traceback
}

var CulpritStyle = "1;4"

func (te *Traceback) Pprint(w io.Writer, sourceIndent string) {
	if te.Begin == -1 {
		fmt.Fprintf(w, "%s, unknown position", te.Name)
		return
	} else if te.Begin < 0 || te.End > len(te.Source) || te.Begin > te.End {
		fmt.Fprintf(w, "%s, invalid position %d-%d", te.Name, te.Begin, te.End)
		return
	}

	before, culprit, after := bca(te.Source, te.Begin, te.End)
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

	if beginLine == endLine {
		fmt.Fprintf(w, "%s, line %d:\n", te.Name, beginLine)
	} else {
		fmt.Fprintf(w, "%s, line %d-%d:\n", te.Name, beginLine, endLine)
	}

	fmt.Fprintf(w, "%s%s", sourceIndent, lineBefore)

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
