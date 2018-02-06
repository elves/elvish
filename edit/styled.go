package edit

import (
	"github.com/elves/elvish/edit/ui"
)

func styled(text, style string) *ui.Styled {
	return &ui.Styled{text, ui.StylesFromString(style)}
}
