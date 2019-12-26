// Package histwalk implements the history walking addon.
package histwalk

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/cli/term"
)

var ErrHistWalkInactive = errors.New("the histwalk addon is not active")

// Config keeps the configuration for the histwalk addon.
type Config struct {
	// Keybinding.
	Binding cli.Handler
	// The history walker.
	Walker histutil.Walker
}

type widget struct {
	app cli.App
	Config
}

func (w *widget) Render(width, height int) *term.Buffer {
	content := cli.ModeLine(
		fmt.Sprintf(" HISTORY #%d ", w.Walker.CurrentSeq()), false)
	buf := term.NewBufferBuilder(width).WriteStyled(content).Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *widget) Handle(event term.Event) bool {
	handled := w.Binding.Handle(event)
	if handled {
		return true
	}
	Accept(w.app)
	return w.app.CodeArea().Handle(event)
}

func (w *widget) Focus() bool { return false }

func (w *widget) onWalk() {
	prefix := w.Walker.Prefix()
	w.app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
		s.Pending = cli.PendingCode{
			From: len(prefix), To: len(s.Buffer.Content),
			Content: w.Walker.CurrentCmd()[len(prefix):],
		}
	})
}

// Start starts the histwalk addon.
func Start(app cli.App, cfg Config) {
	if cfg.Walker == nil {
		app.Notify("no history walker")
		return
	}
	if cfg.Binding == nil {
		cfg.Binding = cli.DummyHandler{}
	}
	walker := cfg.Walker
	err := walker.Prev()
	if err != nil {
		app.Notify(err.Error())
		return
	}
	w := widget{app: app, Config: cfg}
	w.onWalk()
	app.MutateState(func(s *cli.State) { s.Addon = &w })
	app.Redraw()
}

// Prev walks to the previous entry in history. It returns ErrHistWalkInactive
// if the histwalk addon is not active, and histutil.ErrEndOfHistory if it would
// go over the end.
func Prev(app cli.App) error {
	return walk(app, func(w *widget) error { return w.Walker.Prev() })
}

// Next walks to the next entry in history. It returns ErrHistWalkInactive if
// the histwalk addon is not active, and histutil.ErrEndOfHistory if it would go
// over the end.
func Next(app cli.App) error {
	return walk(app, func(w *widget) error { return w.Walker.Next() })
}

// Close closes the histwalk addon. It does nothing if the histwalk addon is not
// active.
func Close(app cli.App) {
	if closeAddon(app) {
		app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
			s.Pending = cli.PendingCode{}
		})
	}
}

// Accept closes the histwalk addon, accepting the current shown command. It does
// nothing if the histwalk addon is not active.
func Accept(app cli.App) {
	if closeAddon(app) {
		app.CodeArea().MutateState(func(s *cli.CodeAreaState) {
			s.ApplyPending()
		})
	}
}

func closeAddon(app cli.App) bool {
	var closed bool
	app.MutateState(func(s *cli.State) {
		if _, ok := s.Addon.(*widget); !ok {
			return
		}
		s.Addon = nil
		closed = true
	})
	return closed
}

func walk(app cli.App, f func(*widget) error) error {
	w, ok := getWidget(app)
	if !ok {
		return ErrHistWalkInactive
	}
	err := f(w)
	if err == nil {
		w.onWalk()
	}
	return err
}

func getWidget(app cli.App) (*widget, bool) {
	w, ok := app.CopyState().Addon.(*widget)
	return w, ok
}
