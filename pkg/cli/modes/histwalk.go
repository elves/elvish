package modes

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
	// Update buffer with current entry. Always returns a nil error.
	Accept() error
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
	app        cli.App
	attachedTo tk.CodeArea
	cursor     histutil.Cursor
	HistwalkSpec
}

func (w *histwalk) Render(width, height int) *term.Buffer {
	buf := w.render(width)
	buf.TrimToLines(0, height)
	return buf
}

func (w *histwalk) MaxHeight(width, height int) int {
	return len(w.render(width).Lines)
}

func (w *histwalk) render(width int) *term.Buffer {
	cmd, _ := w.cursor.Get()
	content := modeLine(fmt.Sprintf(" HISTORY #%d ", cmd.Seq), false)
	return term.NewBufferBuilder(width).WriteStyled(content).Buffer()
}

func (w *histwalk) Handle(event term.Event) bool {
	handled := w.Bindings.Handle(w, event)
	if handled {
		return true
	}
	w.attachedTo.MutateState((*tk.CodeAreaState).ApplyPending)
	w.app.PopAddon()
	return w.attachedTo.Handle(event)
}

func (w *histwalk) Focus() bool { return false }

var errNoHistoryStore = errors.New("no history store")

// NewHistwalk creates a new Histwalk mode.
func NewHistwalk(app cli.App, cfg HistwalkSpec) (Histwalk, error) {
	codeArea, err := FocusedCodeArea(app)
	if err != nil {
		return nil, err
	}
	if cfg.Store == nil {
		return nil, errNoHistoryStore
	}
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	cursor := cfg.Store.Cursor(cfg.Prefix)
	cursor.Prev()
	if _, err := cursor.Get(); err != nil {
		return nil, err
	}
	w := histwalk{app: app, attachedTo: codeArea, HistwalkSpec: cfg, cursor: cursor}
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

func (w *histwalk) Dismiss() {
	w.attachedTo.MutateState(func(s *tk.CodeAreaState) { s.Pending = tk.PendingCode{} })
}

func (w *histwalk) updatePending() {
	cmd, _ := w.cursor.Get()
	w.attachedTo.MutateState(func(s *tk.CodeAreaState) {
		s.Pending = tk.PendingCode{
			From: len(w.Prefix), To: len(s.Buffer.Content),
			Content: cmd.Text[len(w.Prefix):],
		}
	})
}

func (w *histwalk) Accept() error {
	w.attachedTo.MutateState((*tk.CodeAreaState).ApplyPending)
	w.app.PopAddon()
	return nil
}
