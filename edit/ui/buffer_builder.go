package ui

import (
	"strings"

	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

// BufferBuilder supports building of Buffer.
type BufferBuilder struct {
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

// NewBufferBuilder makes a new BufferBuilder, initially with one empty line.
func NewBufferBuilder(width int) *BufferBuilder {
	return &BufferBuilder{Width: width, Lines: [][]Cell{make([]Cell, 0, width)}}
}

func (bb *BufferBuilder) Cursor() Pos {
	return Pos{len(bb.Lines) - 1, bb.Col}
}

// Buffer returns a Buffer built by the BufferBuilder.
func (bb *BufferBuilder) Buffer() *Buffer {
	return &Buffer{bb.Width, bb.Lines, bb.Dot}
}

func (bb *BufferBuilder) SetIndent(indent int) *BufferBuilder {
	bb.Indent = indent
	return bb
}

func (bb *BufferBuilder) SetEagerWrap(v bool) *BufferBuilder {
	bb.EagerWrap = v
	return bb
}

func (bb *BufferBuilder) SetLines(lines ...[]Cell) *BufferBuilder {
	bb.Lines = lines
	bb.Col = CellsWidth(lines[len(lines)-1])
	return bb
}

func (bb *BufferBuilder) SetDot(dot Pos) *BufferBuilder {
	bb.Dot = dot
	return bb
}

func (bb *BufferBuilder) SetDotHere() *BufferBuilder {
	return bb.SetDot(bb.Cursor())
}

func (bb *BufferBuilder) appendLine() {
	bb.Lines = append(bb.Lines, make([]Cell, 0, bb.Width))
	bb.Col = 0
}

func (bb *BufferBuilder) appendCell(c Cell) {
	n := len(bb.Lines)
	bb.Lines[n-1] = append(bb.Lines[n-1], c)
	bb.Col += util.Wcswidth(c.Text)
}

// Newline starts a newline.
func (bb *BufferBuilder) Newline() *BufferBuilder {
	bb.appendLine()

	if bb.Indent > 0 {
		for i := 0; i < bb.Indent; i++ {
			bb.appendCell(Cell{Text: " "})
		}
	}

	return bb
}

var styleForControlChar = Styles{"inverse"}

// WriteRuneSGR writes a single rune to a buffer with an SGR style, wrapping the
// line when needed. If the rune is a control character, it will be written
// using the caret notation (like ^X) and gets the additional style of
// styleForControlChar.
func (bb *BufferBuilder) WriteRuneSGR(r rune, style string) *BufferBuilder {
	if r == '\n' {
		bb.Newline()
		return bb
	}
	c := Cell{string(r), style}
	if r < 0x20 || r == 0x7f {
		if style != "" {
			style = style + ";" + styleForControlChar.String()
		} else {
			style = styleForControlChar.String()
		}
		c = Cell{"^" + string(r^0x40), style}
	}

	if bb.Col+util.Wcswidth(c.Text) > bb.Width {
		bb.Newline()
		bb.appendCell(c)
	} else {
		bb.appendCell(c)
		if bb.Col == bb.Width && bb.EagerWrap {
			bb.Newline()
		}
	}
	return bb
}

// Write is equivalent to calling WriteStyled with styled.MakeText(text,
// style...).
func (bb *BufferBuilder) Write(text string, styles ...string) *BufferBuilder {
	return bb.WriteStyled(styled.MakeText(text, styles...))
}

// WriteMarkedLines is equivalent to calling WriteStyled with
// styled.MarkLines(args...).
func (bb *BufferBuilder) WriteMarkedLines(args ...interface{}) *BufferBuilder {
	return bb.WriteStyled(styled.MarkLines(args...))
}

// WriteStringSGR writes a string to a buffer with a SGR style.
func (bb *BufferBuilder) WriteStringSGR(text, style string) *BufferBuilder {
	for _, r := range text {
		bb.WriteRuneSGR(r, style)
	}
	return bb
}

// WriteSpacesSGR writes w spaces with an SGR style.
func (bb *BufferBuilder) WriteSpacesSGR(w int, style string) *BufferBuilder {
	return bb.WriteStringSGR(strings.Repeat(" ", w), style)
}

// WriteLegacyStyleds writes a styled text.
func (bb *BufferBuilder) WriteStyled(t styled.Text) *BufferBuilder {
	return bb.WriteLegacyStyleds(FromNewStyledText(t))
}

// WriteLegacyStyleds writes a slice of (legacy) styled structs.
func (bb *BufferBuilder) WriteLegacyStyleds(ss []*Styled) *BufferBuilder {
	for _, s := range ss {
		bb.WriteStringSGR(s.Text, s.Styles.String())
	}
	return bb
}

// Extend adds all lines from b2 to the bottom of this buffer. If moveDot is
// true, the dot is updated to match the dot of b2.
//
// TODO(xiaq): Remove this method.
func (bb *BufferBuilder) Extend(b2 *Buffer, moveDot bool) *BufferBuilder {
	if b2 != nil && b2.Lines != nil {
		if moveDot {
			bb.Dot.Line = b2.Dot.Line + len(bb.Lines)
			bb.Dot.Col = b2.Dot.Col
		}
		bb.Lines = append(bb.Lines, b2.Lines...)
		bb.Col = b2.Col()
	}
	return bb
}

// ExtendRight extends bb to the right. It pads each line in bb to be at least of
// width w and appends the corresponding line in b2 to it, making new lines in bb
// when b2 has more lines than bb.
// BUG(xiaq): after calling ExtendRight, the widths of some lines can exceed
// bb.width.
//
// TODO(xiaq): Remove this method.
func (bb *BufferBuilder) ExtendRight(b2 *Buffer, w int) *BufferBuilder {
	i := 0
	for ; i < len(bb.Lines) && i < len(b2.Lines); i++ {
		if w0 := CellsWidth(bb.Lines[i]); w0 < w {
			bb.Lines[i] = append(bb.Lines[i], makeSpacing(w-w0)...)
		}
		bb.Lines[i] = append(bb.Lines[i], b2.Lines[i]...)
	}
	for ; i < len(b2.Lines); i++ {
		row := append(makeSpacing(w), b2.Lines[i]...)
		bb.Lines = append(bb.Lines, row)
	}
	bb.Col = CellsWidth(bb.Lines[len(bb.Lines)-1])
	return bb
}

func makeSpacing(n int) []Cell {
	s := make([]Cell, n)
	for i := 0; i < n; i++ {
		s[i].Text = " "
	}
	return s
}
