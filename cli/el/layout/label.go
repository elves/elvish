package layout

import (
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Label is a Renderer that writes out a text.
type Label struct {
	Content styled.Text
}

// Render shows the content. If the given box is too small, the text is cropped.
func (l Label) Render(width, height int) *ui.Buffer {
	// TODO: Optimize by stopping as soon as $height rows are written.
	bb := ui.NewBufferBuilder(width)
	bb.WriteStyled(l.Content)
	b := bb.Buffer()
	b.TrimToLines(0, height)
	return b
}

// Handle always returns false.
func (l Label) Handle(event term.Event) bool {
	return false
}
