// Package mode contains utilities for implementing modes.
//
// Modes are just pre-packaged ways to use widgets to achieve a specific task.
// Subpackages of this package contain mode implementations.
package mode

import "src.elv.sh/pkg/ui"

// Line returns a text styled as a modeline.
func Line(content string, space bool) ui.Text {
	t := ui.T(content, ui.Bold, ui.FgWhite, ui.BgMagenta)
	if space {
		t = ui.Concat(t, ui.T(" "))
	}
	return t
}

// Prompt returns a callback suitable as the prompt in the codearea of an addon.
func Prompt(content string, space bool) func() ui.Text {
	p := Line(content, space)
	return func() ui.Text { return p }
}
