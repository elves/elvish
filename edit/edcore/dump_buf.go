package edcore

import (
	"fmt"
	"html"
	"strings"

	"github.com/elves/elvish/eval"
)

func (ed *editor) dumpBuf(fm *eval.Frame) {
	out := fm.OutputFile()
	buf := ed.writer.CurrentBuffer()
	for _, line := range buf.Lines {
		style := ""
		openedSpan := false
		for _, c := range line {
			if c.Style != style {
				if openedSpan {
					fmt.Fprint(out, "</span>")
				}
				var classes []string
				for _, c := range strings.Split(c.Style, ";") {
					classes = append(classes, "sgr-"+c)
				}
				fmt.Fprintf(out,
					`<span class="%s">`, strings.Join(classes, " "))
				style = c.Style
				openedSpan = true
			}
			fmt.Fprintf(out, "%s", html.EscapeString(c.Text))
		}
		if openedSpan {
			fmt.Fprint(out, "</span>")
		}
		fmt.Fprint(out, "\n")
	}
}
