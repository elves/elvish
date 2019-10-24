package layout

import "github.com/elves/elvish/styled"

// ModeLine returns a text styled as a modeline.
func ModeLine(content string, space bool) styled.Text {
	t := styled.MakeText(content, "bold", "lightgray", "bg-magenta")
	if space {
		t = t.ConcatText(styled.Plain(" "))
	}
	return t
}

// ModePrompt returns a callback suitable as the prompt in the codearea of a
// combobox.
func ModePrompt(content string, space bool) func() styled.Text {
	p := ModeLine(content, space)
	return func() styled.Text { return p }
}
