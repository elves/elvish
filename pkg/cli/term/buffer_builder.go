package term

import (
	"strings"

	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
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
	Lines Lines
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

func (bb *BufferBuilder) setDot(dot Pos) *BufferBuilder {
	bb.Dot = dot
	return bb
}

func (bb *BufferBuilder) SetDotHere() *BufferBuilder {
	return bb.setDot(bb.Cursor())
}

func (bb *BufferBuilder) appendLine() {
	bb.Lines = append(bb.Lines, make([]Cell, 0, bb.Width))
	bb.Col = 0
}

func (bb *BufferBuilder) appendCell(c Cell) {
	n := len(bb.Lines)
	bb.Lines[n-1] = append(bb.Lines[n-1], c)
	bb.Col += wcwidth.Of(c.Text)
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
		// Always show control characters in reverse video.
		if style != "" {
			style = style + ";7"
		} else {
			style = "7"
		}
		c = Cell{"^" + string(r^0x40), style}
	}

	if bb.Col+wcwidth.Of(c.Text) > bb.Width {
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

// Write is equivalent to calling WriteStyled with ui.T(text, style...).
func (bb *BufferBuilder) Write(text string, ts ...ui.Styling) *BufferBuilder {
	return bb.WriteStyled(ui.T(text, ts...))
}

// WriteSpaces writes w spaces with the given styles.
func (bb *BufferBuilder) WriteSpaces(w int, ts ...ui.Styling) *BufferBuilder {
	return bb.Write(strings.Repeat(" ", w), ts...)
}

// DotHere is a special argument to MarkLines to mark the position of the dot.
var DotHere = struct{ x struct{} }{}

// MarkLines is like calling WriteStyled with ui.MarkLines(args...), but accepts
// an additional special parameter DotHere to mark the position of the dot.
func (bb *BufferBuilder) MarkLines(args ...any) *BufferBuilder {
	for i, arg := range args {
		if arg == DotHere {
			return bb.WriteStyled(ui.MarkLines(args[:i]...)).
				SetDotHere().WriteStyled(ui.MarkLines(args[i+1:]...))
		}
	}
	return bb.WriteStyled(ui.MarkLines(args...))
}

// WriteStringSGR writes a string to a buffer with a SGR style.
func (bb *BufferBuilder) WriteStringSGR(text, style string) *BufferBuilder {
	for _, r := range text {
		bb.WriteRuneSGR(r, style)
	}
	return bb
}

// WriteStyled writes a styled text.
func (bb *BufferBuilder) WriteStyled(t ui.Text) *BufferBuilder {
	for _, seg := range t {
		bb.WriteStringSGR(seg.Text, seg.Style.SGR())
	}
	return bb
}

func makeSpacing(n int) []Cell {
	s := make([]Cell, n)
	for i := 0; i < n; i++ {
		s[i].Text = " "
	}
	return s
}
