package ui

import (
	"github.com/elves/elvish/util"
)

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

// NewStringRenderer returns a Renderer that shows the given string.
func NewStringRenderer(s string) Renderer {
	return stringRenderer{s}
}

type stringRenderer struct {
	s string
}

func (r stringRenderer) Render(bb *BufferBuilder) {
	bb.WriteString(util.TrimWcwidth(r.s, bb.Width), "")
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

// NewModeLineWithScrollBarRenderer returns a Renderer for a mode line with a
// horizontal scroll bar. The base argument should be a Renderer built with
// NewModeLineRenderer; the arguments n, low and high describes the range of the
// scroll bar.
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
