// Package mode implements modes, which are widgets tailored for a specific
// task.
package mode

import "src.elv.sh/pkg/ui"

// Returns text styled as a modeline.
func modeLine(content string, space bool) ui.Text {
	t := ui.T(content, ui.Bold, ui.FgWhite, ui.BgMagenta)
	if space {
		t = ui.Concat(t, ui.T(" "))
	}
	return t
}

func modePrompt(content string, space bool) func() ui.Text {
	p := modeLine(content, space)
	return func() ui.Text { return p }
}

// Prompt returns a callback suitable as the prompt in the codearea of a
// mode widget.
var Prompt = modePrompt
