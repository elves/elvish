// Package histwalk implements the history walking addon.
package histwalk

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

var ErrHistWalkInactive = errors.New("the histwalk addon is not active")

// Config keeps the configuration for the histwalk addon.
type Config struct {
	// Keybinding.
	Binding el.Handler
	// The history walker.
	Walker histutil.Walker
}

type widget struct {
	binding el.Handler
	walker  histutil.Walker
	onWalk  func()
}

func (w *widget) init() {
	if w.binding == nil {
		w.binding = el.DummyHandler{}
	}
	if w.onWalk == nil {
		w.onWalk = func() {}
	}
}

func (w *widget) Render(width, height int) *ui.Buffer {
	w.init()
	content := layout.ModeLine(
		fmt.Sprintf(" HISTORY #%d ", w.walker.CurrentSeq()), false)
	buf := ui.NewBufferBuilder(width).WriteStyled(content).Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *widget) Handle(event term.Event) bool {
	w.init()
	return w.binding.Handle(event)
}

// Start starts the histwalk addon.
func Start(app cli.App, cfg Config) {
	if cfg.Walker == nil {
		app.Notify("no history walker")
		return
	}
	walker := cfg.Walker
	prefix := walker.Prefix()
	walker.Prev()
	w := widget{binding: cfg.Binding, walker: walker}
	w.onWalk = func() {
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.PendingCode = codearea.PendingCode{
				From: len(prefix), To: len(s.CodeBuffer.Content),
				Content: walker.CurrentCmd()[len(prefix):],
			}
		})
	}
	w.init()
	w.onWalk()
	app.MutateState(func(s *cli.State) { s.Listing = &w })
	app.Redraw()
}

// Prev walks to the previous entry in history. It returns ErrHistWalkInactive
// if the histwalk addon is not active.
func Prev(app cli.App) error {
	return walk(app, func(w *widget) error { return w.walker.Prev() })
}

// Next walks to the next entry in history. It returns ErrHistWalkInactive if
// the histwalk addon is not active.
func Next(app cli.App) error {
	return walk(app, func(w *widget) error { return w.walker.Next() })
}

// Close closes the histwalk addon. It does nothing if the histwalk addon is not
// active.
func Close(app cli.App) {
	app.MutateState(func(s *cli.State) {
		if _, ok := s.Listing.(*widget); !ok {
			return
		}
		s.Listing = nil
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.PendingCode = codearea.PendingCode{}
		})
	})
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
	w, ok := app.CopyState().Listing.(*widget)
	return w, ok
}
