package edit

import (
	"fmt"
	"html"
	"strings"

	"github.com/elves/elvish/eval"
)

func _dumpBuf(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	out := ec.OutputFile()
	buf := ec.Editor.(*Editor).writer.oldBuf
	for i, line := range buf.lines {
		if i > 0 {
			fmt.Fprint(out, "<br>")
		}
		style := ""
		openedSpan := false
		for _, c := range line {
			if c.style != style {
				if openedSpan {
					fmt.Fprint(out, "</span>")
				}
				var classes []string
				for _, c := range strings.Split(c.style, ";") {
					classes = append(classes, "sgr-"+c)
				}
				fmt.Fprintf(out,
					`<span class="%s">`, strings.Join(classes, " "))
				style = c.style
				openedSpan = true
			}
			fmt.Fprintf(out, "%s", html.EscapeString(c.string))
		}
		if openedSpan {
			fmt.Fprint(out, "</span>")
		}
	}
}
