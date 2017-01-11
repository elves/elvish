package edit

import (
	"strings"

	"github.com/elves/elvish/util"
)

// cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type cell struct {
	rune
	width byte
	style string
}

// pos is the position within a buffer.
type pos struct {
	line, col int
}

var invalidPos = pos{-1, -1}

func lineWidth(cs []cell) int {
	w := 0
	for _, c := range cs {
		w += int(c.width)
	}
	return w
}

func makeSpacing(n int) []cell {
	s := make([]cell, n)
	for i := 0; i < n; i++ {
		s[i].rune = ' '
		s[i].width = 1
	}
	return s
}

// buffer reflects a continuous range of lines on the terminal. The Unix
// terminal API provides only awkward ways of querying the terminal buffer, so
// we keep an internal reflection and do one-way synchronizations (buffer ->
// terminal, and not the other way around). This requires us to exactly match
// the terminal's idea of the width of characters (wcwidth) and where to
// insert soft carriage returns, so there could be bugs.
type buffer struct {
	width, col, indent int
	newlineWhenFull    bool
	cells              [][]cell // cells reflect len(cells) lines on the terminal.
	dot                pos      // dot is what the user perceives as the cursor.
}

func newBuffer(width int) *buffer {
	return &buffer{width: width, cells: [][]cell{make([]cell, 0, width)}}
}

func (b *buffer) appendCell(c cell) {
	n := len(b.cells)
	b.cells[n-1] = append(b.cells[n-1], c)
	b.col += int(c.width)
}

func (b *buffer) appendLine() {
	b.cells = append(b.cells, make([]cell, 0, b.width))
	b.col = 0
}

func (b *buffer) newline() {
	b.appendLine()

	if b.indent > 0 {
		for i := 0; i < b.indent; i++ {
			b.appendCell(cell{rune: ' ', width: 1})
		}
	}
}

func (b *buffer) extend(b2 *buffer, moveDot bool) {
	if b2 != nil && b2.cells != nil {
		if moveDot {
			b.dot.line = b2.dot.line + len(b.cells)
			b.dot.col = b2.dot.col
		}
		b.cells = append(b.cells, b2.cells...)
		b.col = b2.col
	}
}

// extendHorizontal extends b horizontally. It pads each line in b to be at
// least of width w and appends the corresponding line in b2 to it, making new
// lines in b when b2 has more lines than b.
func (b *buffer) extendHorizontal(b2 *buffer, w int) {
	i := 0
	for ; i < len(b.cells) && i < len(b2.cells); i++ {
		if w0 := lineWidth(b.cells[i]); w0 < w {
			b.cells[i] = append(b.cells[i], makeSpacing(w-w0)...)
		}
		b.cells[i] = append(b.cells[i], b2.cells[i]...)
	}
	for ; i < len(b2.cells); i++ {
		row := append(makeSpacing(w), b2.cells[i]...)
		b.cells = append(b.cells, row)
	}
}

// write appends a single rune to a buffer.
func (b *buffer) write(r rune, style string) {
	if r == '\n' {
		b.newline()
		return
	} else if r < 0x20 || r == 0x7f {
		// BUG(xiaq): buffer.write drops ASCII control characters silently
		return
	}
	wd := util.Wcwidth(r)
	c := cell{r, byte(wd), style}

	if b.col+wd > b.width {
		b.newline()
		b.appendCell(c)
	} else {
		b.appendCell(c)
		if b.col == b.width && b.newlineWhenFull {
			b.newline()
		}
	}
}

func (b *buffer) writes(s string, style string) {
	for _, r := range s {
		b.write(r, style)
	}
}

func (b *buffer) writeStyled(s *styled) {
	b.writes(s.text, s.styles.String())
}

func (b *buffer) writeStyleds(ss []*styled) {
	for _, s := range ss {
		b.writeStyled(s)
	}
}

func (b *buffer) writePadding(w int, style string) {
	b.writes(strings.Repeat(" ", w), style)
}

func (b *buffer) line() int {
	return len(b.cells) - 1
}

func (b *buffer) cursor() pos {
	return pos{len(b.cells) - 1, b.col}
}

func (b *buffer) trimToLines(low, high int) {
	for i := 0; i < low; i++ {
		b.cells[i] = nil
	}
	for i := high; i < len(b.cells); i++ {
		b.cells[i] = nil
	}
	b.cells = b.cells[low:high]
	b.dot.line -= low
}

func widthOfCells(cells []cell) int {
	w := 0
	for _, c := range cells {
		w += int(c.width)
	}
	return w
}

func lines(bufs ...*buffer) (l int) {
	for _, buf := range bufs {
		if buf != nil {
			l += len(buf.cells)
		}
	}
	return
}

func compareRows(r1, r2 []cell) (bool, int) {
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
