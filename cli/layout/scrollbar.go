package layout

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// VScrollbarContainer is a Renderer consisting of content and a vertical
// scrollbar on the right.
type VScrollbarContainer struct {
	Content   clitypes.Renderer
	Scrollbar VScrollbar
}

func (v VScrollbarContainer) Render(width, height int) *ui.Buffer {
	bb := ui.NewBufferBuilder(width)
	bufContent := v.Content.Render(width-1, height)
	bufScrollbar := v.Scrollbar.Render(1, height)
	bb.ExtendRight(bufContent, 0)
	bb.ExtendRight(bufScrollbar, width-1)
	return bb.Buffer()
}

// VScrollbar is a Renderer for a vertical scrollbar.
type VScrollbar struct {
	Total int
	Low   int
	High  int
}

var (
	scrollbarThumb  = styled.MakeText(" ", "magenta", "inverse")
	scrollbarTrough = styled.MakeText("│", "magenta")
)

func (v VScrollbar) Render(width, height int) *ui.Buffer {
	posLow, posHigh := findScrollInterval(v.Total, v.Low, v.High, height)
	bb := ui.NewBufferBuilder(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			bb.Newline()
		}
		if posLow <= i && i < posHigh {
			bb.WriteStyled(scrollbarThumb)
		} else {
			bb.WriteStyled(scrollbarTrough)
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
