package clicore

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/edit/ui"
)

// Renders notes. This does not respect height so that overflow notes end up in
// the scrollback buffer.
func renderNotes(notes []string, width int) *ui.Buffer {
	bb := ui.NewBufferBuilder(width)
	for i, note := range notes {
		if i > 0 {
			bb.Newline()
		}
		bb.WritePlain(note)
	}
	return bb.Buffer()
}

type mainRenderer struct {
	codeArea clitypes.Renderer
	listing  clitypes.Renderer
}

// Renders the codearea, and uses the rest of the height for the listing.
func (r mainRenderer) Render(width, height int) *ui.Buffer {
	buf := r.codeArea.Render(width, height)
	if r.listing != nil && len(buf.Lines) < height {
		bufListing := r.listing.Render(width, height-len(buf.Lines))
		buf.Extend(bufListing, true)
	}
	return buf
}
