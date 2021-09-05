package tk

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// Label is a Renderer that writes out a text.
type Label struct {
	Content ui.Text
}

// Render shows the content. If the given box is too small, the text is cropped.
func (l Label) Render(width, height int) *term.Buffer {
	// TODO: Optimize by stopping as soon as $height rows are written.
	b := l.render(width)
	b.TrimToLines(0, height)
	return b
}

// MaxHeight returns the maximum height the Label can take when rendering within
// a bound box.
func (l Label) MaxHeight(width, height int) int {
	return len(l.render(width).Lines)
}

func (l Label) render(width int) *term.Buffer {
	return term.NewBufferBuilder(width).WriteStyled(l.Content).Buffer()
}

// Handle always returns false.
func (l Label) Handle(event term.Event) bool {
	return false
}
