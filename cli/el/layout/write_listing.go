package layout

import (
	"fmt"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Selected is a special value in the argument to WriteListing, signalling that
// the argument before it is the selected line.
const Selected = "<- selected"

// WriteListing is a unit test helper that emulates the rendering of a "listing"
// type addon. Among the lines arguments, the value "<- selected" is special and
// signals that the argument before it is the selected line.
func WriteListing(b *ui.BufferBuilder, name, filter string, lines ...string) {
	b.WriteStyled(ModeLine(name, true)).
		WritePlain(filter).SetDotHere()
	for i, line := range lines {
		switch {
		case line == Selected:
			continue
		case i < len(lines)-1 && lines[i+1] == Selected:
			b.Newline()
			padded := fmt.Sprintf("%-*s", b.Width, line)
			b.WriteStyled(styled.MakeText(padded, "inverse"))
		default:
			b.Newline()
			b.WritePlain(line)
		}
	}
}
