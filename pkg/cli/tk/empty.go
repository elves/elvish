package tk

import (
	"src.elv.sh/pkg/cli/term"
)

// Empty is an empty widget.
type Empty struct{}

// Render shows nothing, although the resulting Buffer still occupies one line.
func (Empty) Render(width, height int) *term.Buffer {
	return term.NewBufferBuilder(width).Buffer()
}

// MaxHeight returns 1, since this widget always occupies one line.
func (Empty) MaxHeight(width, height int) int {
	return 1
}

// Handle always returns false.
func (Empty) Handle(event term.Event) bool {
	return false
}
