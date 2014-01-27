package edit

import (
	"bytes"
	"fmt"
	"github.com/xiaq/das/edit/tty"
	"github.com/xiaq/das/util"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// cell is an indivisible unit on the screen. It is not necessarily 1 column
// wide.
type cell struct {
	rune
	width byte
	attr  string
}

// pos is the position within a buffer.
type pos struct {
	line, col int
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

func (b *buffer) extend(b2 *buffer) {
	if b2 != nil && b2.cells != nil {
		b.cells = append(b.cells, b2.cells...)
		b.col = b2.col
	}
}

// write appends a single rune to a buffer.
func (b *buffer) write(r rune, attr string) {
	if r == '\n' {
		b.newline()
		return
	} else if !unicode.IsPrint(r) {
		// XXX unprintable runes are dropped silently
		return
	}
	wd := wcwidth(r)
	c := cell{r, byte(wd), attr}

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

func (b *buffer) writes(s string, attr string) {
	for _, r := range s {
		b.write(r, attr)
	}
}

func (b *buffer) writePadding(w int, attr string) {
	b.writes(strings.Repeat(" ", w), attr)
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

// writer is the part of an Editor responsible for keeping the status of and
// updating the screen.
type writer struct {
	file   *os.File
	oldBuf *buffer
}

func newWriter(f *os.File) *writer {
	writer := &writer{file: f, oldBuf: newBuffer(0)}
	return writer
}

// deltaPos calculates the escape sequence needed to move the cursor from one
// position to another.
func deltaPos(from, to pos) []byte {
	buf := new(bytes.Buffer)
	if from.line < to.line {
		// move down
		buf.WriteString(fmt.Sprintf("\033[%dB", to.line-from.line))
	} else if from.line > to.line {
		// move up
		buf.WriteString(fmt.Sprintf("\033[%dA", from.line-to.line))
	}
	if from.col < to.col {
		// move right
		buf.WriteString(fmt.Sprintf("\033[%dC", to.col-from.col))
	} else if from.col > to.col {
		// move left
		buf.WriteString(fmt.Sprintf("\033[%dD", from.col-to.col))
	}
	return buf.Bytes()
}

// commitBuffer updates the terminal display to reflect current buffer.
// TODO Instead of erasing w.oldBuf entirely and then draw buf, compute a
// delta between w.oldBuf and buf
func (w *writer) commitBuffer(buf *buffer) error {
	bytesBuf := new(bytes.Buffer)

	pLine := w.oldBuf.dot.line
	if pLine > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", pLine)
	}
	bytesBuf.WriteString("\r\033[J")

	attr := ""
	for i, line := range buf.cells {
		if i > 0 {
			bytesBuf.WriteString("\n")
		}
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
	cursor := buf.cursor()
	if cursor.col == buf.width {
		cursor.col--
	}
	bytesBuf.Write(deltaPos(cursor, buf.dot))

	_, err := w.file.Write(bytesBuf.Bytes())
	if err != nil {
		return err
	}

	w.oldBuf = buf
	return nil
}

func lines(bufs ...*buffer) (l int) {
	for _, buf := range bufs {
		if buf != nil {
			l += len(buf.cells)
		}
	}
	return
}

// refresh redraws the line editor. The dot is passed as an index into text;
// the corresponding position will be calculated.
func (w *writer) refresh(bs *editorState) error {
	winsize := tty.GetWinsize(int(w.file.Fd()))
	width, height := int(winsize.Col), int(winsize.Row)

	var bufLine, bufMode, bufTips, bufCompletion, buf *buffer
	// bufLine
	b := newBuffer(width)
	bufLine = b

	b.newlineWhenFull = true

	b.writes(bs.prompt, attrForPrompt)

	if b.line() == 0 && b.col*2 < b.width {
		b.indent = b.col
	}

	// i keeps track of number of bytes written.
	i := 0
	if bs.dot == 0 {
		b.dot = b.cursor()
	}

	comp := bs.completion
	var suppress = false
	for _, token := range bs.tokens {
		for _, r := range token.Val {
			if suppress && i < comp.end {
				// Silence the part that is being completed
			} else {
				b.write(r, attrForType[token.Typ])
			}
			i += utf8.RuneLen(r)
			if comp != nil && comp.current != -1 && i == comp.start {
				// Put the current candidate and instruct text up to comp.end
				// to be suppressed. The cursor should be placed correctly
				// (i.e. right after the candidate)
				for _, part := range comp.candidates[comp.current].parts {
					attr := attrForType[comp.typ]
					if part.completed {
						attr += attrForCompleted
					}
					b.writes(part.text, attr)
				}
				suppress = true
			}
			if bs.dot == i {
				b.dot = b.cursor()
			}
		}
	}

	// Write rprompt
	padding := b.width - b.col - wcwidths(bs.rprompt)
	if padding >= 1 {
		b.newlineWhenFull = false
		b.writePadding(padding, "")
		b.writes(bs.rprompt, attrForRprompt)
	}

	// bufMode
	if bs.mode != ModeInsert {
		b := newBuffer(width)
		bufMode = b
		switch bs.mode {
		case ModeCommand:
			b.writes(trimWcwidth("-- COMMAND --", width), attrForMode)
		case ModeCompleting:
			b.writes(trimWcwidth("-- COMPLETING --", width), attrForMode)
		}
	}

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(bs.tips) > 0 {
		b := newBuffer(width)
		bufTips = b
		b.writes(trimWcwidth(strings.Join(bs.tips, ", "), width), attrForTip)
	}

	// bufCompletion
	if comp != nil {
		b := newBuffer(width)
		bufCompletion = b
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

		cols := (b.width + colMargin) / (colWidth + colMargin)
		if cols == 0 {
			cols = 1
		}
		lines := util.CeilDiv(len(cands), cols)
		bs.completionLines = lines

		for i := 0; i < lines; i++ {
			if i > 0 {
				b.newline()
			}
			for j := 0; j < cols; j++ {
				k := j*lines + i
				if k >= len(cands) {
					continue
				}
				var attr string
				if k == comp.current {
					// XXX abuse dot to represent line of current completion
					attr = attrForCurrentCompletion
					b.dot.line = i
				}
				text := cands[k].text
				b.writes(text, attr)
				b.writePadding(colWidth-wcwidths(text), attr)
				b.writePadding(colMargin, "")
			}
		}
	}

	// Trim lines to fit in screen
	switch {
	case height >= lines(bufLine, bufMode, bufTips, bufCompletion):
		// No need to trim.
	case height >= lines(bufLine, bufMode, bufTips):
		h := height - lines(bufLine, bufMode, bufTips)
		// Trim bufCompletion to h lines around the current candidate
		lines := len(bufCompletion.cells)
		low := bufCompletion.dot.line - h/2
		high := low + h
		switch {
		case low < 0:
			// Near top of the list, move the window down
			low = 0
			high = low + h
		case high > lines:
			// Near bottom of the list, move the window down
			high = lines
			low = high - h
		}
		bufCompletion.trimToLines(low, high)
	case height >= lines(bufLine, bufTips):
		bufMode, bufCompletion = nil, nil
	case height >= lines(bufLine):
		bufTips, bufMode, bufCompletion = nil, nil, nil
	case height >= 1:
		bufTips, bufMode, bufCompletion = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine+1-height, dotLine+1)
	default:
		bufLine, bufTips, bufMode, bufCompletion = nil, nil, nil, nil
	}

	// Combine buffers (reusing bufLine)
	buf = bufLine
	buf.extend(bufMode)
	buf.extend(bufTips)
	buf.extend(bufCompletion)

	return w.commitBuffer(buf)
}
