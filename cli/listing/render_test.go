package listing

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

func TestStyledLinesRenderer(t *testing.T) {
	renderer := NewStyledTextsRenderer([]styled.Text{
		styled.Plain("a"),
		styled.Plain("b\nc"),
	})

	wantBuf := ui.NewBufferBuilder(10).WriteString("a", "").Newline().
		WriteString("b", "").Newline().WriteString("c", "").Buffer()
	if buf := ui.Render(renderer, 10); !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Render(renderer, 10) = %v, want %v", buf, wantBuf)
	}
}
