package edit

import (
	"strings"

	"github.com/elves/elvish/util"
)

// cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type cell struct {
	string
	width byte
	style string
}

// Pos is the position within a buffer.
type Pos struct {
	line, col int
}

var invalidPos = Pos{-1, -1}

// cellsWidth returns the total width of a slice of cells.
func cellsWidth(cs []cell) int {
	w := 0
	for _, c := range cs {
		w += int(c.width)
	}
	return w
}

func makeSpacing(n int) []cell {
	s := make([]cell, n)
	for i := 0; i < n; i++ {
		s[i].string = " "
		s[i].width = 1
	}
	return s
}

func compareCells(r1, r2 []cell) (bool, int) {
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

// buffer reflects a continuous range of lines on the terminal.
//
// The Unix terminal API provides only awkward ways of querying the terminal
// buffer, so we keep an internal reflection and do one-way synchronizations
// (buffer -> terminal, and not the other way around). This requires us to
// exactly match the terminal's idea of the width of characters (wcwidth) and
// where to insert soft carriage returns, so there could be bugs.
type buffer struct {
	width, col, indent int
	// eagerWrap controls whether to wrap line as soon as the cursor reaches the
	// right edge of the terminal. This is not often desirable as it creates
	// unneessary line breaks, but is is useful when echoing the user input.
	// will otherwise
	eagerWrap bool
	// lines the content of the buffer.
	lines [][]cell
	dot   Pos // dot is what the user perceives as the cursor.
}

// newBuffer builds a new buffer, with one empty line.
func newBuffer(width int) *buffer {
	return &buffer{width: width, lines: [][]cell{make([]cell, 0, width)}}
}

func (b *buffer) setIndent(indent int) *buffer {
	b.indent = indent
	return b
}

func (b *buffer) setEagerWrap(v bool) *buffer {
	b.eagerWrap = v
	return b
}

func (b *buffer) setLines(lines ...[]cell) *buffer {
	b.lines = lines
	b.col = cellsWidth(lines[len(lines)-1])
	return b
}

func (b *buffer) setDot(dot Pos) *buffer {
	b.dot = dot
	return b
}

func (b *buffer) cursor() Pos {
	return Pos{len(b.lines) - 1, b.col}
}

func buffersHeight(bufs ...*buffer) (l int) {
	for _, buf := range bufs {
		if buf != nil {
			l += len(buf.lines)
		}
	}
	return
}

// Low level buffer mutations.

func (b *buffer) appendLine() {
	b.lines = append(b.lines, make([]cell, 0, b.width))
	b.col = 0
}

func (b *buffer) appendCell(c cell) {
	n := len(b.lines)
	b.lines[n-1] = append(b.lines[n-1], c)
	b.col += int(c.width)
}

// High-level buffer mutations.

func (b *buffer) newline() {
	b.appendLine()

	if b.indent > 0 {
		for i := 0; i < b.indent; i++ {
			b.appendCell(cell{string: " ", width: 1})
		}
	}
}

// write appends a single rune to a buffer, wrapping the line when needed. If
// the rune is a control character, it will be written using the caret notation
// (like ^X) and gets the additional style of styleForControlChar.
func (b *buffer) write(r rune, style string) {
	if r == '\n' {
		b.newline()
		return
	}
	wd := util.Wcwidth(r)
	c := cell{string(r), byte(wd), style}
	if r < 0x20 || r == 0x7f {
		wd = 2
		if style != "" {
			style = style + ";" + styleForControlChar.String()
		} else {
			style = styleForControlChar.String()
		}
		c = cell{"^" + string(r^0x40), 2, style}
	}

	if b.col+wd > b.width {
		b.newline()
		b.appendCell(c)
	} else {
		b.appendCell(c)
		if b.col == b.width && b.eagerWrap {
			b.newline()
		}
	}
}

// writes appends every rune of a string to a buffer, all with the same style.
func (b *buffer) writes(text, style string) {
	for _, r := range text {
		b.write(r, style)
	}
}

// writePadding writes w spaces.
func (b *buffer) writePadding(w int, style string) {
	b.writes(strings.Repeat(" ", w), style)
}

// writeStyleds writes a slice of styled structs.
func (b *buffer) writeStyleds(ss []*styled) {
	for _, s := range ss {
		b.writes(s.text, s.styles.String())
	}
}

// trimToLines trims a buffer to the lines [low, high).
func (b *buffer) trimToLines(low, high int) {
	for i := 0; i < low; i++ {
		b.lines[i] = nil
	}
	for i := high; i < len(b.lines); i++ {
		b.lines[i] = nil
	}
	b.lines = b.lines[low:high]
	b.dot.line -= low
	if b.dot.line < 0 {
		b.dot.line = 0
	}
}

func (b *buffer) extend(b2 *buffer, moveDot bool) {
	if b2 != nil && b2.lines != nil {
		if moveDot {
			b.dot.line = b2.dot.line + len(b.lines)
			b.dot.col = b2.dot.col
		}
		b.lines = append(b.lines, b2.lines...)
		b.col = b2.col
	}
}

// extendRight extends b to the right. It pads each line in b to be at least of
// width w and appends the corresponding line in b2 to it, making new lines in b
// when b2 has more lines than b.
// BUG(xiaq): after calling extendRight, the widths of some lines can exceed
// b.width.
func (b *buffer) extendRight(b2 *buffer, w int) {
	i := 0
	for ; i < len(b.lines) && i < len(b2.lines); i++ {
		if w0 := cellsWidth(b.lines[i]); w0 < w {
			b.lines[i] = append(b.lines[i], makeSpacing(w-w0)...)
		}
		b.lines[i] = append(b.lines[i], b2.lines[i]...)
	}
	for ; i < len(b2.lines); i++ {
		row := append(makeSpacing(w), b2.lines[i]...)
		b.lines = append(b.lines, row)
	}
	b.col = cellsWidth(b.lines[len(b.lines)-1])
}
