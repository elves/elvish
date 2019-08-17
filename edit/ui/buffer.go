package ui

import "github.com/elves/elvish/util"

// Cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type Cell struct {
	Text  string
	Width byte
	Style string
}

// C constructs a Cell.
func C(text, style string) Cell {
	return Cell{text, byte(util.Wcswidth(text)), style}
}

// Pos is the position within a buffer.
type Pos struct {
	Line, Col int
}

// CellsWidth returns the total width of a Cell slice.
func CellsWidth(cs []Cell) int {
	w := 0
	for _, c := range cs {
		w += int(c.Width)
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
