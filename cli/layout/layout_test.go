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
		"Label",
		Label{styled.Plain("label")},
		10, 24,
		ui.NewBufferBuilder(10).WritePlain("label"),
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
