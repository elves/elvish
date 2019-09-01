package layout

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// CroppedLines is a Renderer that writes multiple lines of text, cropping each
// line to suit the width.
type CroppedLines struct {
	Lines []styled.Text
}

// Render shows all the lines, cropping each line to suit the width as well as
// cropping extra lines.
func (c CroppedLines) Render(width, height int) *ui.Buffer {
	lines := c.Lines
	if len(lines) > height {
		lines = lines[:height]
	}
	bb := ui.NewBufferBuilder(width)
	for i, line := range lines {
		if i > 0 {
			bb.Newline()
		}
		bb.WriteStyled(line.TrimWcwidth(width))
	}
	return bb.Buffer()
}
