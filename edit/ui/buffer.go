package ui

import (
	"strings"

	"github.com/elves/elvish/util"
)

// Cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type Cell struct {
	Text  string
	Width byte
	Style string
}

// Pos is the position within a buffer.
type Pos struct {
	Line, Col int
}

var invalidPos = Pos{-1, -1}

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
	Width, Col, Indent int
	// EagerWrap controls whether to wrap line as soon as the cursor reaches the
	// right edge of the terminal. This is not often desirable as it creates
	// unneessary line breaks, but is is useful when echoing the user input.
	// will otherwise
	EagerWrap bool
	// Lines the content of the buffer.
	Lines [][]Cell
	// Dot is what the user perceives as the cursor.
	Dot Pos
}

// NewBuffer builds a new buffer, with one empty line.
func NewBuffer(width int) *Buffer {
	return &Buffer{Width: width, Lines: [][]Cell{make([]Cell, 0, width)}}
}

func (b *Buffer) SetIndent(indent int) *Buffer {
	b.Indent = indent
	return b
}

func (b *Buffer) SetEagerWrap(v bool) *Buffer {
	b.EagerWrap = v
	return b
}

func (b *Buffer) SetLines(lines ...[]Cell) *Buffer {
	b.Lines = lines
	b.Col = CellsWidth(lines[len(lines)-1])
	return b
}

func (b *Buffer) SetDot(dot Pos) *Buffer {
	b.Dot = dot
	return b
}

// Cursor returns the current position of the cursor.
func (b *Buffer) Cursor() Pos {
	return Pos{len(b.Lines) - 1, b.Col}
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

// Low level buffer mutations.

func (b *Buffer) appendLine() {
	b.Lines = append(b.Lines, make([]Cell, 0, b.Width))
	b.Col = 0
}

func (b *Buffer) appendCell(c Cell) {
	n := len(b.Lines)
	b.Lines[n-1] = append(b.Lines[n-1], c)
	b.Col += int(c.Width)
}

// High-level buffer mutations.

// Newline starts a newline.
func (b *Buffer) Newline() {
	b.appendLine()

	if b.Indent > 0 {
		for i := 0; i < b.Indent; i++ {
			b.appendCell(Cell{Text: " ", Width: 1})
		}
	}
}

var styleForControlChar = Styles{"inverse"}

// Write writes a single rune to a buffer, wrapping the line when needed. If the
// rune is a control character, it will be written using the caret notation
// (like ^X) and gets the additional style of styleForControlChar.
func (b *Buffer) Write(r rune, style string) {
	if r == '\n' {
		b.Newline()
		return
	}
	wd := util.Wcwidth(r)
	c := Cell{string(r), byte(wd), style}
	if r < 0x20 || r == 0x7f {
		wd = 2
		if style != "" {
			style = style + ";" + styleForControlChar.String()
		} else {
			style = styleForControlChar.String()
		}
		c = Cell{"^" + string(r^0x40), 2, style}
	}

	if b.Col+wd > b.Width {
		b.Newline()
		b.appendCell(c)
	} else {
		b.appendCell(c)
		if b.Col == b.Width && b.EagerWrap {
			b.Newline()
		}
	}
}

// WriteString writes a string to a buffer, with one style.
func (b *Buffer) WriteString(text, style string) {
	for _, r := range text {
		b.Write(r, style)
	}
}

// WriteSpaces writes w spaces.
func (b *Buffer) WriteSpaces(w int, style string) {
	b.WriteString(strings.Repeat(" ", w), style)
}

// WriteStyleds writes a slice of styled structs.
func (b *Buffer) WriteStyleds(ss []*Styled) {
	for _, s := range ss {
		b.WriteString(s.Text, s.Styles.String())
	}
}

// TrimToLines trims a buffer to the lines [low, high).
func (b *Buffer) TrimToLines(low, high int) {
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
		b.Col = b2.Col
	}
}

// ExtendRight extends b to the right. It pads each line in b to be at least of
// width w and appends the corresponding line in b2 to it, making new lines in b
// when b2 has more lines than b.
// BUG(xiaq): after calling ExtendRight, the widths of some lines can exceed
// b.width.
func (b *Buffer) ExtendRight(b2 *Buffer, w int) {
	i := 0
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
	b.Col = CellsWidth(b.Lines[len(b.Lines)-1])
}

func makeSpacing(n int) []Cell {
	s := make([]Cell, n)
	for i := 0; i < n; i++ {
		s[i].Text = " "
		s[i].Width = 1
	}
	return s
}
