package cliedit

import (
	"fmt"
	"html"
	"strings"

	"github.com/elves/elvish/cli/term"
)

// TODO(xiaq): Move this into the ui package.

func bufToHTML(b *term.Buffer) string {
	var sb strings.Builder
	for _, line := range b.Lines {
		style := ""
		openedSpan := false
		for _, c := range line {
			if c.Style != style {
				if openedSpan {
					sb.WriteString("</span>")
				}
				if c.Style == "" {
					openedSpan = false
				} else {
					var classes []string
					for _, c := range strings.Split(c.Style, ";") {
						classes = append(classes, "sgr-"+c)
					}
					fmt.Fprintf(&sb,
						`<span class="%s">`, strings.Join(classes, " "))
					openedSpan = true
				}
				style = c.Style
			}
			fmt.Fprintf(&sb, "%s", html.EscapeString(c.Text))
		}
		if openedSpan {
			sb.WriteString("</span>")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
