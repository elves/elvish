// Package mode implements modes, which are widgets tailored for a specific
// task.
package mode

import "src.elv.sh/pkg/ui"

// ModeLine returns a text styled as a modeline.
func ModeLine(content string, space bool) ui.Text {
	t := ui.T(content, ui.Bold, ui.FgWhite, ui.BgMagenta)
	if space {
		t = ui.Concat(t, ui.T(" "))
	}
	return t
}

// ModePrompt returns a callback suitable as the prompt in the codearea of a
// mode widget.
func ModePrompt(content string, space bool) func() ui.Text {
	p := ModeLine(content, space)
	return func() ui.Text { return p }
}
