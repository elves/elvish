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
		ui.NewBufferBuilder(10).WritePlain("label"),
	},
	{
		"Label cropping",
		Label{styled.Plain("label")},
		4, 1,
		ui.NewBufferBuilder(4).WritePlain("labe"),
	},
	{
		"CroppedLines showing all",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
		}},
		10, 24,
		ui.NewBufferBuilder(10).WritePlain("line 1").
			Newline().WritePlain("line 2"),
	},
	{
		"CroppedLines cropping horizontally",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
		}},
		4, 24,
		ui.NewBufferBuilder(4).WritePlain("line").
			Newline().WritePlain("line"),
	},
	{
		"CroppedLines cropping vertically",
		CroppedLines{Lines: []styled.Text{
			styled.Plain("line 1"),
			styled.Plain("line 2"),
			styled.Plain("line 3"),
		}},
		10, 2,
		ui.NewBufferBuilder(10).WritePlain("line 1").
			Newline().WritePlain("line 2"),
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
