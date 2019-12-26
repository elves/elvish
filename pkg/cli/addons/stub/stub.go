// Package stub implements the stub addon, a general-purpose addon that shows a
// modeline and supports pluggable binding.
package stub

import (
	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/term"
)

// Config keeps the configuration for the stub addon.
type Config struct {
	// Keybinding.
	Binding cli.Handler
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
		WriteStyled(cli.ModeLine(w.Name, false)).SetDotHere().Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *widget) Handle(event term.Event) bool {
	return w.Binding.Handle(event)
}

func (w *widget) Focus() bool {
	return w.Config.Focus
}

// Start starts the stub addon.
func Start(app cli.App, cfg Config) {
	if cfg.Binding == nil {
		cfg.Binding = cli.DummyHandler{}
	}
	w := widget{cfg}
	app.MutateState(func(s *cli.State) { s.Addon = &w })
	app.Redraw()
}
