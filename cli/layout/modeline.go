package layout

import "github.com/elves/elvish/styled"

// ModePrompt returns a styled text that is suitable as the prompt in the codearea
// of a combobox.
func ModePrompt(content string, space bool) styled.Text {
	p := styled.MakeText(content, "bold", "lightgray", "bg-magenta")
	if space {
		p = p.ConcatText(styled.Plain(" "))
	}
	return p
}
