package layout

import (
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
)

// Empty is an empty widget.
type Empty struct{}

// Render shows nothing, although the resulting Buffer still occupies one line.
func (Empty) Render(width, height int) *ui.Buffer {
	return ui.NewBufferBuilder(width).Buffer()
}

// Handle always returns false.
func (Empty) Handle(event term.Event) bool {
	return false
}
