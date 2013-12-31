package edit

import (
	"os"
	"fmt"
	"bytes"
	"unicode"
	"unicode/utf8"
	"./tty"
	"../parse"
	"../util"
)

// cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type cell struct {
	rune
	width byte
	attr string
}

// pos is the position within a buffer.
type pos struct {
	line, col int
}

// buffer is an internal reflection of the last few lines of the terminal, the
// part the line editor is concerned with.
type buffer struct {
	cells [][]cell // cells reflect the last len(cells) lines of the terminal.
	dot pos // dot is what the user perceives as the cursor.
}

func newBuffer(w int) *buffer {
	return &buffer{cells: [][]cell{make([]cell, 0, w)}}
}

func (b *buffer) appendCell(c cell) {
	n := len(b.cells)
	b.cells[n-1] = append(b.cells[n-1], c)
}

func (b *buffer) appendLine(w int) {
	b.cells = append(b.cells, make([]cell, 0, w))
}

// writer is the part of an Editor responsible for keeping the status of and
// updating the screen.
type writer struct {
	file *os.File
	oldBuf, buf *buffer
	// Fields below are used when refreshing.
	width, indent int
	cursor pos
	currentAttr string
}

func newWriter(f *os.File) *writer {
	writer := &writer{file: f, oldBuf: newBuffer(0)}
	return writer
}

func (w *writer) startBuffer() {
	fd := int(w.file.Fd())
	w.width = int(tty.GetWinsize(fd).Col)
	w.indent = 0
	w.cursor = pos{}
	w.currentAttr = ""
	w.buf = newBuffer(w.width)
}

// deltaPos calculates the escape sequence needed to move the cursor from one
// position to another.
func deltaPos(from, to pos) []byte {
	buf := new(bytes.Buffer)
	if from.line < to.line {
		// move down
		buf.WriteString(fmt.Sprintf("\033[%dB", to.line - from.line))
	} else if from.line > to.line {
		// move up
		buf.WriteString(fmt.Sprintf("\033[%dA", from.line - to.line))
	}
	if from.col < to.col {
		// move right
		buf.WriteString(fmt.Sprintf("\033[%dC", to.col - from.col))
	} else if from.col > to.col {
		// move left
		buf.WriteString(fmt.Sprintf("\033[%dD", from.col - to.col))
	}
	return buf.Bytes()
}

// commitBuffer updates the terminal display to reflect current buffer.
// TODO Instead of erasing w.oldBuf entirely and then draw w.buf, compute a
// delta between w.oldBuf and w.buf
func (w *writer) commitBuffer() error {
	bytesBuf := new(bytes.Buffer)

	pLine := w.oldBuf.dot.line
	if pLine > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", pLine)
	}
	bytesBuf.WriteString("\r\033[J")

	attr := ""
	for _, line := range w.buf.cells {
		for _, c := range line {
			if c.width > 0 && c.attr != attr {
				fmt.Fprintf(bytesBuf, "\033[m\033[%sm", c.attr)
				attr = c.attr
			}
			bytesBuf.WriteString(string(c.rune))
		}
	}
	if attr != "" {
		bytesBuf.WriteString("\033[m")
	}
	bytesBuf.Write(deltaPos(w.cursor, w.buf.dot))

	_, err := w.file.Write(bytesBuf.Bytes())
	if err != nil {
		return err
	}

	w.oldBuf = w.buf
	return nil
}

func (w *writer) appendToLine(c cell) {
	w.buf.appendCell(c)
	w.cursor.col += int(c.width)
}

func (w *writer) newline() {
	w.buf.appendCell(cell{rune: '\n'})
	w.buf.appendLine(w.width)

	w.cursor.line++
	w.cursor.col = 0
	if w.indent > 0 {
		for i := 0; i < w.indent; i++ {
			w.appendToLine(cell{rune: ' ', width: 1})
		}
	}
}

// write appends a single rune to w.buf.
func (w *writer) write(r rune) {
	if r == '\n' {
		w.newline()
		return
	} else if !unicode.IsPrint(r) {
		// XXX unprintable runes are dropped silently
		return
	}
	wd := wcwidth(r)
	c := cell{r, byte(wd), w.currentAttr}

	if w.cursor.col + wd > w.width {
		w.newline()
		w.appendToLine(c)
	} else if w.cursor.col + wd == w.width {
		w.appendToLine(c)
		w.newline()
	} else {
		w.appendToLine(c)
	}
}

// refresh redraws the line editor. The dot is passed as an index into text;
// the corresponding position will be calculated.
func (w *writer) refresh(prompt string, tokens []parse.Item, tip string, comp *completion, dot int) error {
	w.startBuffer()

	for _, r := range prompt {
		w.write(r)
	}

	if w.cursor.col * 2 < w.width {
		w.indent = w.cursor.col
	}

	// i keeps track of number of bytes written.
	i := 0
	if dot == 0 {
		w.buf.dot = w.cursor
	}

	var suppress = false;
	for _, token := range tokens {
		w.currentAttr = attrForType[token.Typ]
		for _, r := range token.Val {
			if suppress && i < comp.end {
				// Silence the part that is being completed
			} else {
				w.write(r)
			}
			i += utf8.RuneLen(r)
			if comp != nil && comp.current != -1 && i == comp.start {
				// Put the current candidate and instruct text up to comp.end
				// to be suppressed. The cursor should be placed correctly
				// (i.e. right after the candidate)
				for _, part := range comp.candidates[comp.current].parts {
					w.currentAttr = attrForType[comp.typ]
					if part.completed {
						w.currentAttr += attrForCompleted
					}
					for _, r := range part.text {
						w.write(r)
					}
				}
				suppress = true;
			}
			if dot == i {
				w.buf.dot = w.cursor
			}
		}
	}

	w.indent = 0
	w.currentAttr = ""
	if len(tip) > 0 {
		w.newline()
		for _, r := range tip {
			w.write(r)
		}
	}

	if comp != nil {
		// Layout candidates in multiple columns
		cands := comp.candidates

		// First decide the shape (# of rows and columns)
		colWidth := 0
		colMargin := 2
		for _, cand := range cands {
			width := wcwidths(cand.text)
			if colWidth < width {
				colWidth = width
			}
		}

		cols := (w.width + colMargin) / (colWidth + colMargin)
		lines := util.CeilDiv(len(cands), cols)

		for i := 0; i < lines; i++ {
			w.newline()
			for j := 0; j < cols; j++ {
				k := j * lines + i
				if k >= len(cands) {
					continue
				}
				if k == comp.current {
					w.currentAttr = attrForCurrentCompletion
				} else {
					w.currentAttr = ""
				}
				for _, r := range cands[k].text {
					w.write(r)
				}
				for l := wcwidths(cands[k].text); l < colWidth; l++ {
					w.write(' ')
				}
				w.currentAttr = ""
				for l := 0; l < colMargin; l++ {
					w.write(' ')
				}
			}
		}
	}

	return w.commitBuffer()
}
