// The highlight program highlights Elvish code fences in Markdown.
package main

import (
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"

	"src.elv.sh/pkg/edit/highlight"
)

func highlightCodeContent(sb *strings.Builder, language string, lines []string) {
	switch language {
	case "elvish", "elvish-bad":
		highlightElvish(sb, lines, language == "elvish-bad")
	case "elvish-transcript":
		highlightElvishTranscript(sb, lines)
	default:
		for _, line := range lines {
			sb.WriteString(html.EscapeString(line))
			sb.WriteByte('\n')
		}
	}
}

var (
	highlighter = highlight.NewHighlighter(highlight.Config{})
	ps1Pattern  = regexp.MustCompile(`^[~/][^ ]*> `)
)

func highlightElvishTranscript(sb *strings.Builder, lines []string) {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if ps1 := ps1Pattern.FindString(line); ps1 != "" {
			elvishLines := []string{line[len(ps1):]}
			ps2 := strings.Repeat(" ", len(ps1))
			for i++; i < len(lines) && strings.HasPrefix(lines[i], ps2); i++ {
				elvishLines = append(elvishLines, lines[i])
			}
			i--
			sb.WriteString(html.EscapeString(ps1))
			highlightElvish(sb, elvishLines, false)
		} else {
			// Write an output line.
			sb.WriteString(html.EscapeString(line))
			sb.WriteByte('\n')
		}
	}
}

func highlightElvish(sb *strings.Builder, lines []string, bad bool) {
	text := strings.Join(lines, "\n") + "\n"

	highlighted, errs := highlighter.Get(text)
	if len(errs) != 0 && !bad {
		log.Printf("parsing %q: %v", text, errs)
	}

	for _, seg := range highlighted {
		var classes []string
		for _, sgrCode := range strings.Split(seg.Style.SGR(), ";") {
			classes = append(classes, "sgr-"+sgrCode)
		}
		jointClass := strings.Join(classes, " ")
		if len(jointClass) > 0 {
			fmt.Fprintf(sb, `<span class="%s">`, jointClass)
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
}
