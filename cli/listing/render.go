package listing

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// NewStyledTextsRenderer returns a Renderer that shows the given styled texts
// joined by newlines, possibly trimmed or padded to fit whatever width is
// available.
//
// NOTE: This is in this package to avoid a cyclic dependency between edit/ui
// and styled. This should be moved into edit/ui as soon as the legacy styled
// types are removed.
func NewStyledTextsRenderer(lines []styled.Text) ui.Renderer {
	return styledTextsRenderer{lines}
}

type styledTextsRenderer struct{ lines []styled.Text }

func (r styledTextsRenderer) Render(bb *ui.BufferBuilder) {
	for i, line := range r.lines {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteLegacyStyleds(line.TrimWcwidth(bb.Width).ToLegacyType())
	}
}
