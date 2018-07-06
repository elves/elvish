package ui

// Renderer wraps the Render method.
type Renderer interface {
	// Render renders onto a Buffer.
	Render(bb *BufferBuilder)
}

// Render creates a new Buffer with the given width, and lets a Renderer render
// onto it.
func Render(r Renderer, width int) *Buffer {
	if r == nil {
		return nil
	}
	bb := NewBufferBuilder(width)
	r.Render(bb)
	return bb.Buffer()
}

// NewModeLineRenderer returns a Renderer for a mode line.
func NewModeLineRenderer(title, filter string) Renderer {
	return modeLineRenderer{title, filter}
}

type modeLineRenderer struct {
	title  string
	filter string
}

func (ml modeLineRenderer) Render(bb *BufferBuilder) {
	bb.WriteString(ml.title, styleForMode.String())
	bb.WriteSpaces(1, "")
	bb.WriteString(ml.filter, styleForFilter.String())
	bb.Dot = bb.Cursor()
}

func NewModeLineWithScrollBarRenderer(base Renderer, n, low, high int) Renderer {
	return &modeLineWithScrollBarRenderer{base, n, low, high}
}

type modeLineWithScrollBarRenderer struct {
	base         Renderer
	n, low, high int
}

func (ml modeLineWithScrollBarRenderer) Render(bb *BufferBuilder) {
	ml.base.Render(bb)

	scrollbarWidth := bb.Width - CellsWidth(bb.Lines[len(bb.Lines)-1]) - 2
	if scrollbarWidth >= 3 {
		bb.WriteSpaces(1, "")
		writeHorizontalScrollbar(bb, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}
