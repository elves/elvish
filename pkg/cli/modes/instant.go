package modes

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// Instant is a mode that executes code whenever it changes and shows the
// result.
type Instant interface {
	tk.Widget
}

// InstantSpec specifies the configuration for the instant mode.
type InstantSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// The function to execute code and returns the output.
	Execute func(code string) ([]string, error)
}

type instant struct {
	InstantSpec
	attachedTo tk.CodeArea
	textView   tk.TextView
	lastCode   string
	lastErr    error
}

func (w *instant) Render(width, height int) *term.Buffer {
	buf := w.render(width, height)
	buf.TrimToLines(0, height)
	return buf
}

func (w *instant) MaxHeight(width, height int) int {
	return len(w.render(width, height).Lines)
}

func (w *instant) render(width, height int) *term.Buffer {
	bb := term.NewBufferBuilder(width).
		WriteStyled(modeLine(" INSTANT ", false)).SetDotHere()
	if w.lastErr != nil {
		bb.Newline().Write(w.lastErr.Error(), ui.FgRed)
	}
	buf := bb.Buffer()
	if len(buf.Lines) < height {
		bufTextView := w.textView.Render(width, height-len(buf.Lines))
		buf.Extend(bufTextView, false)
	}
	return buf
}

func (w *instant) Focus() bool { return false }

func (w *instant) Handle(event term.Event) bool {
	handled := w.Bindings.Handle(w, event)
	if !handled {
		handled = w.attachedTo.Handle(event)
	}
	w.update(false)
	return handled
}

func (w *instant) update(force bool) {
	code := w.attachedTo.CopyState().Buffer.Content
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

var errExecutorIsRequired = errors.New("executor is required")

// NewInstant creates a new instant mode.
func NewInstant(app cli.App, cfg InstantSpec) (Instant, error) {
	codeArea, err := FocusedCodeArea(app)
	if err != nil {
		return nil, err
	}
	if cfg.Execute == nil {
		return nil, errExecutorIsRequired
	}
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	w := instant{
		InstantSpec: cfg,
		attachedTo:  codeArea,
		textView:    tk.NewTextView(tk.TextViewSpec{Scrollable: true}),
	}
	w.update(true)
	return &w, nil
}
