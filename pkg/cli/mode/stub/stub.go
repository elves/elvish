// Package stub implements the stub addon, a general-purpose addon that shows a
// modeline and supports pluggable binding.
package stub

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

// Config keeps the configuration for the stub addon.
type Config struct {
	// Key bindings.
	Bindings tk.Bindings
	// Name to show in the modeline.
	Name string
	// Whether the addon widget gets the focus.
	Focus bool
}

type widget struct {
	Config
}

func (w *widget) Render(width, height int) *term.Buffer {
	buf := term.NewBufferBuilder(width).
		WriteStyled(mode.Line(w.Name, false)).SetDotHere().Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *widget) Handle(event term.Event) bool {
	return w.Bindings.Handle(w, event)
}

func (w *widget) Focus() bool {
	return w.Config.Focus
}

// Start starts the stub addon.
func Start(app cli.App, cfg Config) {
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	w := widget{cfg}
	app.SetAddon(&w, false)
	app.Redraw()
}
