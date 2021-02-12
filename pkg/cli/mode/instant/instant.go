// Package instant implements an addon that executes code whenever it changes
// and shows the result.
package instant

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// Config keeps the configuration for the instant addon.
type Config struct {
	// Keybinding.
	Binding tk.Handler
	// The function to execute code and returns the output.
	Execute func(code string) ([]string, error)
}

type widget struct {
	Config
	app      cli.App
	textView tk.TextView
	lastCode string
	lastErr  error
}

func (w *widget) Render(width, height int) *term.Buffer {
	bb := term.NewBufferBuilder(width).
		WriteStyled(mode.Line(" INSTANT ", false)).SetDotHere()
	if w.lastErr != nil {
		bb.Newline().Write(w.lastErr.Error(), ui.FgRed)
	}
	buf := bb.Buffer()
	if len(buf.Lines) >= height {
		buf.TrimToLines(0, height)
		return buf
	}
	bufTextView := w.textView.Render(width, height-len(buf.Lines))
	buf.Extend(bufTextView, false)
	return buf
}

func (w *widget) Focus() bool { return false }

func (w *widget) Handle(event term.Event) bool {
	handled := w.Binding.Handle(event)
	if !handled {
		codeArea := w.app.CodeArea()
		handled = codeArea.Handle(event)
	}
	w.update(false)
	return handled
}

func (w *widget) update(force bool) {
	code := w.app.CodeArea().CopyState().Buffer.Content
	if code == w.lastCode && !force {
		return
	}
	w.lastCode = code
	output, err := w.Execute(code)
	w.lastErr = err
	if err == nil {
		w.textView.MutateState(func(s *tk.TextViewState) {
			*s = tk.TextViewState{Lines: output, First: 0}
		})
	}
}

// Start starts the addon.
func Start(app cli.App, cfg Config) {
	if cfg.Execute == nil {
		app.Notify("executor is required")
		return
	}
	if cfg.Binding == nil {
		cfg.Binding = tk.DummyHandler{}
	}
	w := widget{
		Config:   cfg,
		app:      app,
		textView: tk.NewTextView(tk.TextViewSpec{Scrollable: true}),
	}
	w.update(true)
	app.SetAddon(&w, false)
	app.Redraw()
}
