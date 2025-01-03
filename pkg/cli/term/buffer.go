package term

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/wcwidth"
)

// Cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type Cell struct {
	Text  string
	Style string
}

// Pos is a line/column position.
type Pos struct {
	Line, Col int
}

// Returns the total width of a Cell slice.
func cellsWidth(cs []Cell) int {
	w := 0
	for _, c := range cs {
		w += wcwidth.Of(c.Text)
	}
	return w
}

// Returns whether two Cell slices are equal, and when they are not, the first
// index at which they differ.
func compareCells(r1, r2 []Cell) (bool, int) {
	for i, c := range r1 {
		if i >= len(r2) || c != r2[i] {
			return false, i
		}
	}
	if len(r1) < len(r2) {
		return false, len(r1)
	}
	return true, 0
}

// Buffer reflects a rectangle area in the terminal, along with a cursor (called
// a "dot" here).
//
// The Unix terminal API provides only awkward ways of querying the terminal
// Buffer, so we keep an internal reflection and do one-way synchronizations
// (Buffer -> terminal, and not the other way around). This requires us to
// exactly match the terminal's idea of the width of characters (wcwidth) and
// where to insert soft carriage returns, so there could be bugs.
type Buffer struct {
	Width int
	// Lines the content of the buffer.
	Lines [][]Cell
	// Dot is what the user perceives as the cursor.
	Dot Pos
}

// Returns the position of the cursor after writing the entire buffer.
func endPos(b *Buffer) Pos {
	return Pos{len(b.Lines) - 1, cellsWidth(b.Lines[len(b.Lines)-1])}
}

// TrimToLines trims a buffer to the lines [low, high).
func (b *Buffer) TrimToLines(low, high int) {
	if low < 0 {
		low = 0
	}
	if high > len(b.Lines) {
		high = len(b.Lines)
	}
	for i := 0; i < low; i++ {
		b.Lines[i] = nil
	}
	for i := high; i < len(b.Lines); i++ {
		b.Lines[i] = nil
	}
	b.Lines = b.Lines[low:high]
	b.Dot.Line -= low
	if b.Dot.Line < 0 {
		b.Dot.Line = 0
	}
}

// ExtendDown extends b downwards, by adding all lines from b2 to the bottom of
// this buffer and setting b.Width to the larger of b.Width and b2.Width. If
// moveDot is true, it also updates b.Dot to match the dot of b2. It returns b
// itself.
func (b *Buffer) ExtendDown(b2 *Buffer, moveDot bool) *Buffer {
	if b2 == nil || b2.Lines == nil {
		return b
	}
	if moveDot {
		b.Dot = Pos{Line: len(b.Lines) + b2.Dot.Line, Col: b2.Dot.Col}
	}
	b.Lines = append(b.Lines, b2.Lines...)
	b.Width = max(b.Width, b2.Width)
	return b
}

// ExtendRight extends b to the right, by padding each line in b to be b.Width
// and appends the corresponding line in b2 to it, making new lines when b2 has
// more lines than bb. If moveDot is true, it also updates b.Dot to match the
// dot of b2. It returns b itself.
func (b *Buffer) ExtendRight(b2 *Buffer, moveDot bool) *Buffer {
	i := 0
	for ; i < len(b.Lines) && i < len(b2.Lines); i++ {
		if w0 := cellsWidth(b.Lines[i]); w0 < b.Width {
			b.Lines[i] = append(b.Lines[i], makeSpacing(b.Width-w0)...)
		}
		b.Lines[i] = append(b.Lines[i], b2.Lines[i]...)
	}
	for ; i < len(b2.Lines); i++ {
		row := append(makeSpacing(b.Width), b2.Lines[i]...)
		b.Lines = append(b.Lines, row)
	}

	if moveDot {
		b.Dot = Pos{Line: b2.Dot.Line, Col: b.Width + b2.Dot.Col}
	}
	b.Width += b2.Width
	return b
}

// Buffer returns itself. This is implemented in analogy with [BufferBuilder],
// so that places that accept either can accept an interface.
func (b *Buffer) Buffer() *Buffer { return b }

// TTYString returns a text representation of the buffer. It uses box drawing
// characters to represent the border of the buffer, and embeds SGR sequences to
// represent the style of the text.
func (b *Buffer) TTYString() string {
	if b == nil {
		return "nil"
	}
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "Width = %d, Dot = (%d, %d)\n", b.Width, b.Dot.Line, b.Dot.Col)
	// Top border
	sb.WriteString("┌" + strings.Repeat("─", b.Width) + "┐\n")
	for _, line := range b.Lines {
		// Left border
		sb.WriteRune('│')
		// Content
		lastStyle := ""
		usedWidth := 0
		for _, cell := range line {
			if cell.Style != lastStyle {
				switch {
				case lastStyle == "":
					sb.WriteString("\033[" + cell.Style + "m")
				case cell.Style == "":
					sb.WriteString("\033[m")
				default:
					sb.WriteString("\033[;" + cell.Style + "m")
				}
				lastStyle = cell.Style
			}
			sb.WriteString(cell.Text)
			usedWidth += wcwidth.Of(cell.Text)
		}
		if lastStyle != "" {
			sb.WriteString("\033[m")
		}
		if usedWidth < b.Width {
			sb.WriteString("$" + strings.Repeat(" ", b.Width-usedWidth-1))
		}
		// Right border and newline
		sb.WriteString("│\n")
	}
	// Bottom border
	sb.WriteString("└" + strings.Repeat("─", b.Width) + "┘\n")
	return sb.String()
}
