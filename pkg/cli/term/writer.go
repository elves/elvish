package term

import (
	"bytes"
	"fmt"
	"io"
)

var logWriterDetail = false

// Writer represents the output to a terminal.
type Writer interface {
	// Buffer returns the current buffer.
	Buffer() *Buffer
	// ResetBuffer resets the current buffer.
	ResetBuffer()
	// UpdateBuffer updates the terminal display to reflect current buffer.
	UpdateBuffer(bufNoti, buf *Buffer, fullRefresh bool) error
	// ClearScreen clears the terminal screen and places the cursor at the top
	// left corner.
	ClearScreen()
	// ShowCursor shows the cursor.
	ShowCursor()
	// HideCursor hides the cursor.
	HideCursor()
}

// writer renders the editor UI.
type writer struct {
	file   io.Writer
	curBuf *Buffer
}

// NewWriter returns a Writer that writes VT100 sequences to the given io.Writer.
func NewWriter(f io.Writer) Writer {
	return &writer{f, &Buffer{}}
}

func (w *writer) Buffer() *Buffer {
	return w.curBuf
}

func (w *writer) ResetBuffer() {
	w.curBuf = &Buffer{}
}

// deltaPos calculates the escape sequence needed to move the cursor from one
// position to another. It use relative movements to move to the destination
// line and absolute movement to move to the destination column.
func deltaPos(from, to Pos) []byte {
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

const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

// UpdateBuffer updates the terminal display to reflect current buffer.
func (w *writer) UpdateBuffer(bufNoti, buf *Buffer, fullRefresh bool) error {
	if buf.Width != w.curBuf.Width && w.curBuf.Lines != nil {
		// Width change, force full refresh
		w.curBuf.Lines = nil
		fullRefresh = true
	}

	bytesBuf := new(bytes.Buffer)

	bytesBuf.WriteString(hideCursor)

	// Rewind cursor
	if pLine := w.curBuf.Dot.Line; pLine > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", pLine)
	}
	bytesBuf.WriteString("\r")

	if fullRefresh {
		// Erase from here. We may be in the top right corner of the screen; if
		// we simply do an erase here, tmux will save the current screen in the
		// scrollback buffer (presumably as a heuristics to detect full-screen
		// applications), but that is not something we want. So we write a space
		// first, and then erase, before rewinding back.
		//
		// Source code for tmux behavior:
		// https://github.com/tmux/tmux/blob/5f5f029e3b3a782dc616778739b2801b00b17c0e/screen-write.c#L1139
		bytesBuf.WriteString(" \033[J\r")
	}

	// style of last written cell.
	style := ""

	switchStyle := func(newstyle string) {
		if newstyle != style {
			fmt.Fprintf(bytesBuf, "\033[0;%sm", newstyle)
			style = newstyle
		}
	}

	writeCells := func(cs []Cell) {
		for _, c := range cs {
			switchStyle(c.Style)
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
		// TODO(xiaq): This is hacky; try to improve it.
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
			if eq, j = CompareCells(line, w.curBuf.Lines[i]); eq {
				continue
			}
		}
		// Move to the first differing column if necessary.
		firstCol := CellsWidth(line[:j])
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
	bytesBuf.WriteString(showCursor)

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

func (w *writer) HideCursor() {
	fmt.Fprint(w.file, hideCursor)
}

func (w *writer) ShowCursor() {
	fmt.Fprint(w.file, showCursor)
}

func (w *writer) ClearScreen() {
	fmt.Fprint(w.file,
		"\033[H",  // move cursor to the top left corner
		"\033[2J", // clear entire buffer
	)
}
