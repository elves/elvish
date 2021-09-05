package modes

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
)

// Stub is a mode that just shows a modeline and keeps the focus on the code
// area. It is mainly useful to apply some special non-default bindings.
type Stub interface {
	tk.Widget
}

// StubSpec specifies the configuration for the stub mode.
type StubSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// Name to show in the modeline.
	Name string
}

type stub struct {
	StubSpec
}

func (w stub) Render(width, height int) *term.Buffer {
	buf := w.render(width)
	buf.TrimToLines(0, height)
	return buf
}

func (w stub) MaxHeight(width, height int) int {
	return len(w.render(width).Lines)
}

func (w stub) render(width int) *term.Buffer {
	return term.NewBufferBuilder(width).
		WriteStyled(modeLine(w.Name, false)).SetDotHere().Buffer()
}

func (w stub) Handle(event term.Event) bool {
	return w.Bindings.Handle(w, event)
}

func (w stub) Focus() bool {
	return false
}

// NewStub creates a new Stub mode.
func NewStub(cfg StubSpec) Stub {
	if cfg.Bindings == nil {
		cfg.Bindings = tk.DummyBindings{}
	}
	return stub{cfg}
}
