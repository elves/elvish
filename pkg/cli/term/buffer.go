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

// CellsWidth returns the total width of a Cell slice.
func CellsWidth(cs []Cell) int {
	w := 0
	for _, c := range cs {
		w += wcwidth.Of(c.Text)
	}
	return w
}

// CompareCells returns whether two Cell slices are equal, and when they are
// not, the first index at which they differ.
func CompareCells(r1, r2 []Cell) (bool, int) {
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

// Buffer reflects a continuous range of lines on the terminal.
//
// The Unix terminal API provides only awkward ways of querying the terminal
// Buffer, so we keep an internal reflection and do one-way synchronizations
// (Buffer -> terminal, and not the other way around). This requires us to
// exactly match the terminal's idea of the width of characters (wcwidth) and
// where to insert soft carriage returns, so there could be bugs.
type Buffer struct {
	Width int
	// Lines the content of the buffer.
	Lines Lines
	// Dot is what the user perceives as the cursor.
	Dot Pos
}

// Lines stores multiple lines.
type Lines [][]Cell

// Line stores a single line.
type Line []Cell

// NewBuffer builds a new buffer, with one empty line.
func NewBuffer(width int) *Buffer {
	return &Buffer{Width: width, Lines: [][]Cell{make([]Cell, 0, width)}}
}

// Col returns the column the cursor is in.
func (b *Buffer) Col() int {
	return CellsWidth(b.Lines[len(b.Lines)-1])
}

// Cursor returns the current position of the cursor.
func (b *Buffer) Cursor() Pos {
	return Pos{len(b.Lines) - 1, b.Col()}
}

// BuffersHeight computes the combined height of a number of buffers.
func BuffersHeight(bufs ...*Buffer) (l int) {
	for _, buf := range bufs {
		if buf != nil {
			l += len(buf.Lines)
		}
	}
	return
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

// Extend adds all lines from b2 to the bottom of this buffer. If moveDot is
// true, the dot is updated to match the dot of b2.
func (b *Buffer) Extend(b2 *Buffer, moveDot bool) {
	if b2 != nil && b2.Lines != nil {
		if moveDot {
			b.Dot.Line = b2.Dot.Line + len(b.Lines)
			b.Dot.Col = b2.Dot.Col
		}
		b.Lines = append(b.Lines, b2.Lines...)
	}
}

// ExtendRight extends bb to the right. It pads each line in b to be b.Width and
// appends the corresponding line in b2 to it, making new lines when b2 has more
// lines than bb.
func (b *Buffer) ExtendRight(b2 *Buffer) {
	i := 0
	w := b.Width
	b.Width += b2.Width
	for ; i < len(b.Lines) && i < len(b2.Lines); i++ {
		if w0 := CellsWidth(b.Lines[i]); w0 < w {
			b.Lines[i] = append(b.Lines[i], makeSpacing(w-w0)...)
		}
		b.Lines[i] = append(b.Lines[i], b2.Lines[i]...)
	}
	for ; i < len(b2.Lines); i++ {
		row := append(makeSpacing(w), b2.Lines[i]...)
		b.Lines = append(b.Lines, row)
	}
}

// Buffer returns itself.
func (b *Buffer) Buffer() *Buffer { return b }

// TTYString returns a string for representing the buffer on the terminal.
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
