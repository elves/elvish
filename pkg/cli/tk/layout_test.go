package tk

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

var layoutRenderTests = []struct {
	name     string
	renderer Renderer
	width    int
	height   int
	wantBuf  *term.BufferBuilder
}{
	{
		"empty widget",
		Empty{},
		10, 24,
		bb(10),
	},
	{
		"Label showing all",
		Label{ui.T("label")},
		10, 24,
		bb(10).Write("label"),
	},
	{
		"Label cropping",
		Label{ui.T("label")},
		4, 1,
		bb(4).Write("labe"),
	},
	{
		"VScrollbar showing full thumb",
		VScrollbar{4, 0, 3},
		10, 2,
		bb(1).WriteStyled(vscrollbarThumb).WriteStyled(vscrollbarThumb),
	},
	{
		"VScrollbar showing thumb in first half",
		VScrollbar{4, 0, 1},
		10, 2,
		bb(1).WriteStyled(vscrollbarThumb).WriteStyled(vscrollbarTrough),
	},
	{
		"VScrollbar showing a minimal 1-size thumb at beginning",
		VScrollbar{4, 0, 0},
		10, 2,
		bb(1).WriteStyled(vscrollbarThumb).WriteStyled(vscrollbarTrough),
	},
	{
		"VScrollbar showing a minimal 1-size thumb at end",
		VScrollbar{4, 3, 3},
		10, 2,
		bb(1).WriteStyled(vscrollbarTrough).WriteStyled(vscrollbarThumb),
	},
	{
		"VScrollbarContainer",
		VScrollbarContainer{Label{ui.T("abcd1234")},
			VScrollbar{4, 0, 1}},
		5, 2,
		bb(5).Write("abcd").WriteStyled(vscrollbarThumb).
			Newline().Write("1234").WriteStyled(vscrollbarTrough),
	},
	{
		"HScrollbar showing full thumb",
		HScrollbar{4, 0, 3},
		2, 10,
		bb(2).WriteStyled(hscrollbarThumb).WriteStyled(hscrollbarThumb),
	},
	{
		"HScrollbar showing thumb in first half",
		HScrollbar{4, 0, 1},
		2, 10,
		bb(2).WriteStyled(hscrollbarThumb).WriteStyled(hscrollbarTrough),
	},
	{
		"HScrollbar showing a minimal 1-size thumb at beginning",
		HScrollbar{4, 0, 0},
		2, 10,
		bb(2).WriteStyled(hscrollbarThumb).WriteStyled(hscrollbarTrough),
	},
	{
		"HScrollbar showing a minimal 1-size thumb at end",
		HScrollbar{4, 3, 3},
		2, 10,
		bb(2).WriteStyled(hscrollbarTrough).WriteStyled(hscrollbarThumb),
	},
}

func TestLayout_Render(t *testing.T) {
	for _, test := range layoutRenderTests {
		t.Run(test.name, func(t *testing.T) {
			buf := test.renderer.Render(test.width, test.height)
			wantBuf := test.wantBuf.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("got buf %v, want %v", buf, wantBuf)
			}
		})
	}
}

var nopHandlers = []Handler{
	Empty{}, Label{ui.T("label")},
}

func TestLayout_Handle(t *testing.T) {
	for _, handler := range nopHandlers {
		handled := handler.Handle(term.K('a'))
		if handled {
			t.Errorf("%v handles event when it shouldn't", handler)
		}
	}
}
