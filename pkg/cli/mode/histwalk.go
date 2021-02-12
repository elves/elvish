package mode

import (
	"errors"
	"fmt"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

// Histwalk is a mode for walking through history.
type Histwalk interface {
	tk.Widget
	// Walk to the previous entry in history.
	Prev() error
	// Walk to the next entry in history.
	Next() error
}

// HistwalkSpec specifies the configuration for the histwalk mode.
type HistwalkSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// History store to walk.
	Store histutil.Store
	// Only walk through items with this prefix.
	Prefix string
}

type histwalk struct {
	app    cli.App
	cursor histutil.Cursor
	HistwalkSpec
}

func (w *histwalk) Render(width, height int) *term.Buffer {
	cmd, _ := w.cursor.Get()
	content := ModeLine(fmt.Sprintf(" HISTORY #%d ", cmd.Seq), false)
	buf := term.NewBufferBuilder(width).WriteStyled(content).Buffer()
	buf.TrimToLines(0, height)
	return buf
}

func (w *histwalk) Handle(event term.Event) bool {
	handled := w.Bindings.Handle(w, event)
	if handled {
		return true
	}
	w.app.SetAddon(nil, true)
	return w.app.CodeArea().Handle(event)
}

func (w *histwalk) Focus() bool { return false }

var errNoHistoryStore = errors.New("no history store")

// NewHistwalk creates a new Histwalk mode.
func NewHistwalk(app cli.App, cfg HistwalkSpec) (Histwalk, error) {
	if cfg.Store == nil {
		return nil, errNoHistoryStore
	}
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	cursor := cfg.Store.Cursor(cfg.Prefix)
	cursor.Prev()
	_, err := cursor.Get()
	if err != nil {
		return nil, err
	}
	w := histwalk{app: app, HistwalkSpec: cfg, cursor: cursor}
	w.updatePending()
	return &w, nil
}

func (w *histwalk) Prev() error {
	return w.walk(histutil.Cursor.Prev, histutil.Cursor.Next)
}

func (w *histwalk) Next() error {
	return w.walk(histutil.Cursor.Next, histutil.Cursor.Prev)
}

func (w *histwalk) walk(f func(histutil.Cursor), undo func(histutil.Cursor)) error {
	f(w.cursor)
	_, err := w.cursor.Get()
	if err == nil {
		w.updatePending()
	} else if err == histutil.ErrEndOfHistory {
		undo(w.cursor)
	}
	return err
}

func (w *histwalk) Close(accept bool) {
	w.app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
		if accept {
			s.ApplyPending()
		} else {
			s.Pending = tk.PendingCode{}
		}
	})
}

func (w *histwalk) updatePending() {
	cmd, _ := w.cursor.Get()
	w.app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
		s.Pending = tk.PendingCode{
			From: len(w.Prefix), To: len(s.Buffer.Content),
			Content: cmd.Text[len(w.Prefix):],
		}
	})
}
