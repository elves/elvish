// Package mode implements modes, which are widgets tailored for a specific
// task.
package modes

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// ErrFocusedWidgetNotCodeArea is returned when an operation requires the
// focused widget to be a code area but it is not.
var ErrFocusedWidgetNotCodeArea = errors.New("focused widget is not a code area")

// FocusedCodeArea returns a CodeArea widget if the currently focused widget is
// a CodeArea. Otherwise it returns the error ErrFocusedWidgetNotCodeArea.
func FocusedCodeArea(a cli.App) (tk.CodeArea, error) {
	if w, ok := a.FocusedWidget().(tk.CodeArea); ok {
		return w, nil
	}
	return nil, ErrFocusedWidgetNotCodeArea
}

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

// ErrorText returns a red "error:" followed by unstyled space and err.Error().
func ErrorText(err error) ui.Text {
	return ui.Concat(ui.T("error:", ui.FgRed), ui.T(" "), ui.T(err.Error()))
}
