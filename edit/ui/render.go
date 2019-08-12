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

// NewStringRenderer returns a Renderer that shows the given strings unstyled,
// trimmed to fit whatever width is available.
func NewStringRenderer(s string) Renderer {
	return NewLinesRenderer(s)
}

// NewLinesRenderer returns a Renderer that shows the given lines unstyled,
// each trimmed to fit whatever width is available.
func NewLinesRenderer(lines ...string) Renderer {
	return linesRenderer{lines}
}

type linesRenderer struct{ lines []string }

func (r linesRenderer) Render(bb *BufferBuilder) {
	for i, line := range r.lines {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteStringSGR(util.TrimWcwidth(line, bb.Width), "")
	}
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
	bb.WriteStringSGR(ml.title, styleForMode.String())
	bb.WriteSpacesSGR(1, "")
	bb.WriteStringSGR(ml.filter, styleForFilter.String())
	bb.SetDotToCursor()
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
		bb.WriteSpacesSGR(1, "")
		writeHorizontalScrollbar(bb, ml.n, ml.low, ml.high, scrollbarWidth)
	}
}

// NewRendererWithVerticalScrollbar returns a Renderer that renders the given
// base plus a vertical scrollbar at the right-hand side.
func NewRendererWithVerticalScrollbar(base Renderer, n, low, high int) Renderer {
	return rendererWithVerticalScrollbar{base, n, low, high}
}

type rendererWithVerticalScrollbar struct {
	base         Renderer
	n, low, high int
}

func (r rendererWithVerticalScrollbar) Render(bb *BufferBuilder) {
	bufBase := Render(r.base, bb.Width-1)
	bb.ExtendRight(bufBase, 0)
	bufScrollbar := renderVerticalScrollbar(r.n, r.low, r.high, len(bufBase.Lines))
	bb.ExtendRight(bufScrollbar, bb.Width-1)
}
