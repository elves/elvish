package elvdoc

import (
	"regexp"
	"strings"

	"src.elv.sh/pkg/edit/highlight"
	"src.elv.sh/pkg/ui"
)

// With an empty highlight.Config, this highlighter does not check for
// compilation errors or non-existent commands.
var highlighter = highlight.NewHighlighter(highlight.Config{})

// HighlightCodeBlock highlights a code block from Markdown. It handles thea
// elvish and elvish-transcript languages. It also removes comment and directive
// lines from elvish-transcript code blocks.
func HighlightCodeBlock(info, code string) ui.Text {
	language, _, _ := strings.Cut(info, " ")
	switch language {
	case "elvish":
		t, _ := highlighter.Get(code)
		return t
	case "elvish-transcript":
		return highlightTranscript(code)
	default:
		return ui.T(code)
	}
}

// Pattern for the prefix of the first line of Elvish code in a transcript.
var ps1Pattern = regexp.MustCompile(`^[~/][^ ]*> `)

// TODO: Ideally this should use the parser in [src.elv.sh/pkg/transcript],
func highlightTranscript(code string) ui.Text {
	var tb ui.TextBuilder
	lines := strings.Split(code, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if ps1 := ps1Pattern.FindString(line); ps1 != "" {
			elvishLines := []string{line[len(ps1):]}
			// Include lines that are indented with the same length of ps1.
			ps2 := strings.Repeat(" ", len(ps1))
			for i++; i < len(lines) && strings.HasPrefix(lines[i], ps2); i++ {
				elvishLines = append(elvishLines, lines[i])
			}
			i--
			highlighted, _ := highlighter.Get(strings.Join(elvishLines, "\n"))
			tb.WriteText(ui.T(ps1))
			tb.WriteText(highlighted)
		} else if strings.HasPrefix(line, "//") {
			// Suppress comment/directive line.
			continue
		} else {
			// Write an output line.
			tb.WriteText(ui.T(line))
		}
		if i < len(lines)-1 {
			tb.WriteText(ui.T("\n"))
		}
	}
	return tb.Text()
}
