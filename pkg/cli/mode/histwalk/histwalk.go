// Package histwalk implements the history walking addon.
package histwalk

import (
	"errors"
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

var ErrHistWalkInactive = errors.New("the histwalk addon is not active")

// Config keeps the configuration for the histwalk addon.
type Config struct {
	// Key bindings.
	Bindings tk.Bindings
	// History store to walk.
	Store histutil.Store
	// Only walk through items with this prefix.
	Prefix string
}

type widget struct {
	app    cli.App
	cursor histutil.Cursor
	Config
}

func (w *widget) Render(width, height int) *term.Buffer {
	cmd, _ := w.cursor.Get()
	content := mode.Line(fmt.Sprintf(" HISTORY #%d ", cmd.Seq), false)
	buf := term.NewBufferBuilder(width).WriteStyled(content).Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *widget) Handle(event term.Event) bool {
	handled := w.Bindings.Handle(w, event)
	if handled {
		return true
	}
	Accept(w.app)
	return w.app.CodeArea().Handle(event)
}

func (w *widget) Focus() bool { return false }

func (w *widget) onWalk() {
	cmd, _ := w.cursor.Get()
	w.app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
		s.Pending = tk.PendingCode{
			From: len(w.Prefix), To: len(s.Buffer.Content),
			Content: cmd.Text[len(w.Prefix):],
		}
	})
}

// Start starts the histwalk addon.
func Start(app cli.App, cfg Config) {
	if cfg.Store == nil {
		app.Notify("no history store")
		return
	}
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	cursor := cfg.Store.Cursor(cfg.Prefix)
	cursor.Prev()
	_, err := cursor.Get()
	if err != nil {
		app.Notify(err.Error())
		return
	}
	w := widget{app: app, Config: cfg, cursor: cursor}
	w.onWalk()
	app.SetAddon(&w, false)
	app.Redraw()
}

// Prev walks to the previous entry in history. It returns ErrHistWalkInactive
// if the histwalk addon is not active, and histutil.ErrEndOfHistory if it would
// go over the end.
func Prev(app cli.App) error {
	return walk(app, histutil.Cursor.Prev, histutil.Cursor.Next)
}

// Next walks to the next entry in history. It returns ErrHistWalkInactive if
// the histwalk addon is not active, and histutil.ErrEndOfHistory if it would go
// over the end.
func Next(app cli.App) error {
	return walk(app, histutil.Cursor.Next, histutil.Cursor.Prev)
}

// Close closes the histwalk addon. It does nothing if the histwalk addon is not
// active.
func Close(app cli.App) {
	if closeAddon(app) {
		app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
			s.Pending = tk.PendingCode{}
		})
	}
}

// Accept closes the histwalk addon, accepting the current shown command. It does
// nothing if the histwalk addon is not active.
func Accept(app cli.App) {
	if closeAddon(app) {
		app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
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

func walk(app cli.App, f func(histutil.Cursor), undo func(histutil.Cursor)) error {
	w, ok := getWidget(app)
	if !ok {
		return ErrHistWalkInactive
	}
	f(w.cursor)
	_, err := w.cursor.Get()
	if err == nil {
		w.onWalk()
	} else if err == histutil.ErrEndOfHistory {
		undo(w.cursor)
	}
	return err
}

func getWidget(app cli.App) (*widget, bool) {
	w, ok := app.CopyState().Addon.(*widget)
	return w, ok
}
