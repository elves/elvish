package layout

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

var bb = ui.NewBufferBuilder

var renderTests = []struct {
	name     string
	renderer clitypes.Renderer
	width    int
	height   int
	wantBuf  *ui.BufferBuilder
}{
	{
		"Label showing all",
		Label{styled.Plain("label")},
		10, 24,
		bb(10).WritePlain("label"),
	},
	{
		"Label cropping",
		Label{styled.Plain("label")},
		4, 1,
		bb(4).WritePlain("labe"),
	},
	{
		"CroppedLines showing all",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
		}},
		10, 24,
		bb(10).WritePlain("line 1").Newline().WritePlain("line 2"),
	},
	{
		"CroppedLines cropping horizontally",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
		}},
		4, 24,
		bb(4).WritePlain("line").Newline().WritePlain("line"),
	},
	{
		"CroppedLines cropping vertically",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
			styled.Plain("line 3"),
		}},
		10, 2,
		bb(10).WritePlain("line 1").Newline().WritePlain("line 2"),
	},
	{
		"VScrollbar showing full thumb",
		VScrollbar{4, 0, 3},
		10, 2,
		bb(1).WriteStyled(scrollbarThumb).WriteStyled(scrollbarThumb),
	},
	{
		"VScrollbar showing thumb in first half",
		VScrollbar{4, 0, 1},
		10, 2,
		bb(1).WriteStyled(scrollbarThumb).WriteStyled(scrollbarTrough),
	},
	{
		"VScrollbar showing a minimal 1-size thumb at beginning",
		VScrollbar{4, 0, 0},
		10, 2,
		bb(1).WriteStyled(scrollbarThumb).WriteStyled(scrollbarTrough),
	},
	{
		"VScrollbar showing a minimal 1-size thumb at end",
		VScrollbar{4, 3, 3},
		10, 2,
		bb(1).WriteStyled(scrollbarTrough).WriteStyled(scrollbarThumb),
	},
	{
		"VScrollbarContainer",
		VScrollbarContainer{Label{styled.Plain("abcd1234")},
			VScrollbar{4, 0, 1}},
		5, 2,
		bb(5).WritePlain("abcd").WriteStyled(scrollbarThumb).
			Newline().WritePlain("1234").WriteStyled(scrollbarTrough),
	},
}

func TestRender(t *testing.T) {
	for _, test := range renderTests {
		t.Run(test.name, func(t *testing.T) {
			buf := test.renderer.Render(test.width, test.height)
			wantBuf := test.wantBuf.Buffer()
			if !reflect.DeepEqual(buf, wantBuf) {
				t.Errorf("got buf %v, want %v", buf, wantBuf)
			}
		})
	}
}
