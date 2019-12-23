package layout

import "github.com/elves/elvish/pkg/ui"

// ModeLine returns a text styled as a modeline.
func ModeLine(content string, space bool) ui.Text {
	t := ui.T(content, ui.Bold, ui.FgWhite, ui.BgMagenta)
	if space {
		t = t.ConcatText(ui.T(" "))
	}
	return t
}

// ModePrompt returns a callback suitable as the prompt in the codearea of a
// combobox.
func ModePrompt(content string, space bool) func() ui.Text {
	p := ModeLine(content, space)
	return func() ui.Text { return p }
}
