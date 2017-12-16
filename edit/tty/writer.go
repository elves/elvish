package tty

import (
	"bytes"
	"fmt"
	"os"

	"github.com/elves/elvish/edit/ui"
)

var logWriterDetail = false

type Writer interface {
	// CurrentBuffer returns the current buffer.
	CurrentBuffer() *ui.Buffer
	// ResetCurrentBuffer resets the current buffer.
	ResetCurrentBuffer()
	// CommitBuffer updates the terminal display to reflect current buffer.
	CommitBuffer(bufNoti, buf *ui.Buffer, fullRefresh bool) error
}

// writer renders the editor UI.
type writer struct {
	file   *os.File
	curBuf *ui.Buffer
}

func NewWriter(f *os.File) Writer {
	return &writer{f, &ui.Buffer{}}
}

// CurrentBuffer returns the current buffer.
func (w *writer) CurrentBuffer() *ui.Buffer {
	return w.curBuf
}

// ResetCurrentBuffer resets the current buffer.
func (w *writer) ResetCurrentBuffer() {
	w.curBuf = &ui.Buffer{}
}

// deltaPos calculates the escape sequence needed to move the cursor from one
// position to another. It use relative movements to move to the destination
// line and absolute movement to move to the destination column.
func deltaPos(from, to ui.Pos) []byte {
	buf := new(bytes.Buffer)
	if from.Line < to.Line {
		// move down
		fmt.Fprintf(buf, "\033[%dB", to.Line-from.Line)
	} else if from.Line > to.Line {
		// move up
		fmt.Fprintf(buf, "\033[%dA", from.Line-to.Line)
	}
	fmt.Fprint(buf, "\r")
	if to.Col > 0 {
		fmt.Fprintf(buf, "\033[%dC", to.Col)
	}
	return buf.Bytes()
}

// CommitBuffer updates the terminal display to reflect current buffer.
func (w *writer) CommitBuffer(bufNoti, buf *ui.Buffer, fullRefresh bool) error {
	if buf.Width != w.curBuf.Width && w.curBuf.Lines != nil {
		// Width change, force full refresh
		w.curBuf.Lines = nil
		fullRefresh = true
	}

	bytesBuf := new(bytes.Buffer)

	// Hide cursor.
	bytesBuf.WriteString("\033[?25l")

	// Rewind cursor
	if pLine := w.curBuf.Dot.Line; pLine > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", pLine)
	}
	bytesBuf.WriteString("\r")

	if fullRefresh {
		// Do an erase.
		bytesBuf.WriteString("\033[J")
	}

	// style of last written cell.
	style := ""

	switchStyle := func(newstyle string) {
		if newstyle != style {
			fmt.Fprintf(bytesBuf, "\033[0;%sm", newstyle)
			style = newstyle
		}
	}

	writeCells := func(cs []ui.Cell) {
		for _, c := range cs {
			if c.Width > 0 {
				switchStyle(c.Style)
			}
			bytesBuf.WriteString(c.Text)
		}
	}

	if bufNoti != nil {
		if logWriterDetail {
			logger.Printf("going to write %d lines of notifications", len(bufNoti.Lines))
		}

		// Write notifications
		for _, line := range bufNoti.Lines {
			writeCells(line)
			switchStyle("")
			bytesBuf.WriteString("\033[K\n")
		}
		// XXX Hacky.
		if len(w.curBuf.Lines) > 0 {
			w.curBuf.Lines = w.curBuf.Lines[1:]
		}
	}

	if logWriterDetail {
		logger.Printf("going to write %d lines, oldBuf had %d", len(buf.Lines), len(w.curBuf.Lines))
	}

	for i, line := range buf.Lines {
		if i > 0 {
			bytesBuf.WriteString("\n")
		}
		var j int // First column where buf and oldBuf differ
		// No need to update current line
		if !fullRefresh && i < len(w.curBuf.Lines) {
			var eq bool
			if eq, j = ui.CompareCells(line, w.curBuf.Lines[i]); eq {
				continue
			}
		}
		// Move to the first differing column if necessary.
		firstCol := ui.CellsWidth(line[:j])
		if firstCol != 0 {
			fmt.Fprintf(bytesBuf, "\033[%dC", firstCol)
		}
		// Erase the rest of the line if necessary.
		if !fullRefresh && i < len(w.curBuf.Lines) && j < len(w.curBuf.Lines[i]) {
			switchStyle("")
			bytesBuf.WriteString("\033[K")
		}
		writeCells(line[j:])
	}
	if len(w.curBuf.Lines) > len(buf.Lines) && !fullRefresh {
		// If the old buffer is higher, erase old content.
		// Note that we cannot simply write \033[J, because if the cursor is
		// just over the last column -- which is precisely the case if we have a
		// rprompt, \033[J will also erase the last column.
		switchStyle("")
		bytesBuf.WriteString("\n\033[J\033[A")
	}
	switchStyle("")
	cursor := buf.Cursor()
	bytesBuf.Write(deltaPos(cursor, buf.Dot))

	// Show cursor.
	bytesBuf.WriteString("\033[?25h")

	if logWriterDetail {
		logger.Printf("going to write %q", bytesBuf.String())
	}

	_, err := w.file.Write(bytesBuf.Bytes())
	if err != nil {
		return err
	}

	w.curBuf = buf
	return nil
}
