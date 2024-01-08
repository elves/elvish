package main

import (
	"fmt"
	"html"
	"strings"

	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/ui"
)

func convertCodeBlock(info, code string) string {
	return textToHTML(elvdoc.HighlightCodeBlock(info, code))
}

func textToHTML(t ui.Text) string {
	var sb strings.Builder
	for _, seg := range t {
		var classes []string
		for _, sgrCode := range seg.Style.SGRValues() {
			classes = append(classes, "sgr-"+sgrCode)
		}
		jointClass := strings.Join(classes, " ")
		if len(jointClass) > 0 {
			fmt.Fprintf(&sb, `<span class="%s">`, jointClass)
		}
		for _, r := range seg.Text {
			if r == '\n' {
				sb.WriteByte('\n')
			} else {
				sb.WriteString(html.EscapeString(string(r)))
			}
		}
		if len(jointClass) > 0 {
			sb.WriteString("</span>")
		}
	}
	return sb.String()
}
