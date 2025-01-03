package term

import (
	"bytes"
	"fmt"
	"io"

	"src.elv.sh/pkg/ui"
)

var logWriterDetail = false

// Writer represents the output to a terminal.
type Writer interface {
	// Buffer returns the current buffer.
	Buffer() *Buffer
	// ResetBuffer resets the current buffer.
	ResetBuffer()
	// UpdateBuffer updates the terminal display to reflect current buffer.
	UpdateBuffer(msg ui.Text, buf *Buffer, fullRefresh bool) error
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
func (w *writer) UpdateBuffer(msg ui.Text, buf *Buffer, fullRefresh bool) error {
	if (buf.Width != w.curBuf.Width && w.curBuf.Lines != nil) || msg != nil {
		// If the width has changed or we have any message to write, we can't do
		// delta rendering meaningfully, so force a full refresh.
		fullRefresh = true
	}

	// Store all the output write in a buffer, so that we only write to the
	// terminal once.
	output := new(bytes.Buffer)

	// Hide cursor at the beginning to minimize flickering.
	output.WriteString(hideCursor)

	// Rewind cursor.
	if pLine := w.curBuf.Dot.Line; pLine > 0 {
		fmt.Fprintf(output, "\033[%dA", pLine)
	}
	output.WriteString("\r")

	if fullRefresh {
		// Erase from here. We may be in the top left corner of the screen; if
		// we simply do an erase here, tmux (and possibly other terminal
		// emulators) will save the current screen in the scrollback buffer,
		// presumably as a heuristic to detect full-screen applications, but
		// that is not something we want.
		//
		// To defeat tmux's heuristic, we write a space, erase, and then rewind.
		//
		// Source code for tmux behavior:
		// https://github.com/tmux/tmux/blob/5f5f029e3b3a782dc616778739b2801b00b17c0e/screen-write.c#L1139
		output.WriteString(" \033[J\r")
	}

	if msg != nil {
		// Write the message with the terminal's line wrapping enabled, for
		// easier copy-pasting by the user.
		output.WriteString("\033[?7h" + msg.VTString() + "\n\033[?7l")
	}

	// style of last written cell.
	style := ""

	switchStyle := func(newstyle string) {
		if newstyle != style {
			fmt.Fprintf(output, "\033[0;%sm", newstyle)
			style = newstyle
		}
	}

	writeCells := func(cs []Cell) {
		for _, c := range cs {
			switchStyle(c.Style)
			output.WriteString(c.Text)
		}
	}

	if logWriterDetail {
		logger.Printf("going to write %d lines, oldBuf had %d", len(buf.Lines), len(w.curBuf.Lines))
	}

	for i, line := range buf.Lines {
		if i > 0 {
			// Move cursor down one line and to the leftmost column. Shorter
			// than "\033[B\r".
			output.WriteString("\n")
		}
		if fullRefresh || i >= len(w.curBuf.Lines) {
			// When doing a full refresh or writing new lines, we have an empty
			// canvas to work with, so just write the current line.
			writeCells(line)
			continue
		}
		// Delta update below.
		eq, j := compareCells(line, w.curBuf.Lines[i])
		if eq {
			// This line hasn't changed
			continue
		}
		// This line has changed, and j is the first differing cell. Move to its
		// corresponding column.
		if firstCol := cellsWidth(line[:j]); firstCol != 0 {
			fmt.Fprintf(output, "\033[%dC", firstCol)
		}
		// Erase the rest of the line; this is not necessary if the old version
		// of the line is a prefix of the current version of the line.
		if j < len(w.curBuf.Lines[i]) {
			switchStyle("")
			output.WriteString("\033[K")
		}
		// Now write the new content.
		writeCells(line[j:])
	}
	if !fullRefresh && len(w.curBuf.Lines) > len(buf.Lines) {
		// If the old buffer is higher, erase old content.
		//
		// Note that we cannot simply write \033[J, because if the cursor is
		// just over the last column -- which is precisely the case if we have a
		// rprompt, \033[J will also erase the last column. Since the old buffer
		// is higher, we know that the \n we write won't create a bogus new
		// line.
		switchStyle("")
		output.WriteString("\n\033[J\033[A")
	}
	switchStyle("")
	cursor := endPos(buf)
	output.Write(deltaPos(cursor, buf.Dot))

	// Show cursor.
	output.WriteString(showCursor)

	if logWriterDetail {
		logger.Printf("going to write %q", output.String())
	}

	_, err := w.file.Write(output.Bytes())
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
