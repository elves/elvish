package ui

// Renderer wraps the Render method.
type Renderer interface {
	// Render renders onto a Buffer.
	Render(b *Buffer)
}

// Render creates a new Buffer with the given width, and lets a Renderer render
// onto it.
func Render(r Renderer, width int) *Buffer {
	if r == nil {
		return nil
	}
	b := NewBuffer(width)
	r.Render(b)
	return b
}

// NewModeLineRenderer returns a Renderer for a mode line.
func NewModeLineRenderer(title, filter string) Renderer {
	return modeLineRenderer{title, filter}
}

type modeLineRenderer struct {
	title  string
	filter string
}

func (ml modeLineRenderer) Render(b *Buffer) {
	b.WriteString(ml.title, styleForMode.String())
	b.WriteSpaces(1, "")
	b.WriteString(ml.filter, styleForFilter.String())
	b.Dot = b.Cursor()
}

func NewModeLineWithScrollBarRenderer(base Renderer, n, low, high int) Renderer {
	return &modeLineWithScrollBarRenderer{base, n, low, high}
}

type modeLineWithScrollBarRenderer struct {
	base         Renderer
	n, low, high int
}

func (ml modeLineWithScrollBarRenderer) Render(b *Buffer) {
	ml.base.Render(b)

	scrollbarWidth := b.Width - CellsWidth(b.Lines[len(b.Lines)-1]) - 2
	if scrollbarWidth >= 3 {
		b.WriteSpaces(1, "")
		writeHorizontalScrollbar(b, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}
