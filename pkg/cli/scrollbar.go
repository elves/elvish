package cli

import (
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

// VScrollbarContainer is a Renderer consisting of content and a vertical
// scrollbar on the right.
type VScrollbarContainer struct {
	Content   Renderer
	Scrollbar VScrollbar
}

func (v VScrollbarContainer) Render(width, height int) *term.Buffer {
	buf := v.Content.Render(width-1, height)
	buf.ExtendRight(v.Scrollbar.Render(1, height))
	return buf
}

// VScrollbar is a Renderer for a vertical scrollbar.
type VScrollbar struct {
	Total int
	Low   int
	High  int
}

var (
	vscrollbarThumb  = ui.T(" ", ui.FgMagenta, ui.Inverse)
	vscrollbarTrough = ui.T("│", ui.FgMagenta)
)

func (v VScrollbar) Render(width, height int) *term.Buffer {
	posLow, posHigh := findScrollInterval(v.Total, v.Low, v.High, height)
	bb := term.NewBufferBuilder(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			bb.Newline()
		}
		if posLow <= i && i < posHigh {
			bb.WriteStyled(vscrollbarThumb)
		} else {
			bb.WriteStyled(vscrollbarTrough)
		}
	}
	return bb.Buffer()
}

// HScrollbar is a Renderer for a horizontal scrollbar.
type HScrollbar struct {
	Total int
	Low   int
	High  int
}

var (
	hscrollbarThumb  = ui.T(" ", ui.FgMagenta, ui.Inverse)
	hscrollbarTrough = ui.T("━", ui.FgMagenta)
)

func (h HScrollbar) Render(width, height int) *term.Buffer {
	posLow, posHigh := findScrollInterval(h.Total, h.Low, h.High, width)
	bb := term.NewBufferBuilder(width)
	for i := 0; i < width; i++ {
		if posLow <= i && i < posHigh {
			bb.WriteStyled(hscrollbarThumb)
		} else {
			bb.WriteStyled(hscrollbarTrough)
		}
	}
	return bb.Buffer()
}

func findScrollInterval(n, low, high, height int) (int, int) {
	f := func(i int) int {
		return int(float64(i)/float64(n)*float64(height) + 0.5)
	}
	scrollLow, scrollHigh := f(low), f(high)
	if scrollLow == scrollHigh {
		if scrollHigh == height {
			scrollLow--
		} else {
			scrollHigh++
		}
	}
	return scrollLow, scrollHigh
}
