package edit

import (
	"bytes"
	"fmt"
	"os"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/sys"
)

var logWriterDetail = false

// Writer renders the editor UI.
type Writer struct {
	file   *os.File
	oldBuf *ui.Buffer
}

func newWriter(f *os.File) *Writer {
	writer := &Writer{file: f, oldBuf: &ui.Buffer{}}
	return writer
}

func (w *Writer) resetOldBuf() {
	w.oldBuf = &ui.Buffer{}
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

// commitBuffer updates the terminal display to reflect current buffer.
// TODO Instead of erasing w.oldBuf entirely and then draw buf, compute a
// delta between w.oldBuf and buf
func (w *Writer) commitBuffer(bufNoti, buf *ui.Buffer, fullRefresh bool) error {
	if buf.Width != w.oldBuf.Width && w.oldBuf.Lines != nil {
		// Width change, force full refresh
		w.oldBuf.Lines = nil
		fullRefresh = true
	}

	bytesBuf := new(bytes.Buffer)

	// Hide cursor.
	bytesBuf.WriteString("\033[?25l")

	// Rewind cursor
	if pLine := w.oldBuf.Dot.Line; pLine > 0 {
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
		if len(w.oldBuf.Lines) > 0 {
			w.oldBuf.Lines = w.oldBuf.Lines[1:]
		}
	}

	if logWriterDetail {
		logger.Printf("going to write %d lines, oldBuf had %d", len(buf.Lines), len(w.oldBuf.Lines))
	}

	for i, line := range buf.Lines {
		if i > 0 {
			bytesBuf.WriteString("\n")
		}
		var j int // First column where buf and oldBuf differ
		// No need to update current line
		if !fullRefresh && i < len(w.oldBuf.Lines) {
			var eq bool
			if eq, j = ui.CompareCells(line, w.oldBuf.Lines[i]); eq {
				continue
			}
		}
		// Move to the first differing column if necessary.
		firstCol := ui.CellsWidth(line[:j])
		if firstCol != 0 {
			fmt.Fprintf(bytesBuf, "\033[%dC", firstCol)
		}
		// Erase the rest of the line if necessary.
		if !fullRefresh && i < len(w.oldBuf.Lines) && j < len(w.oldBuf.Lines[i]) {
			switchStyle("")
			bytesBuf.WriteString("\033[K")
		}
		writeCells(line[j:])
	}
	if len(w.oldBuf.Lines) > len(buf.Lines) && !fullRefresh {
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

	fd := int(w.file.Fd())
	if nonblock, _ := sys.GetNonblock(fd); nonblock {
		sys.SetNonblock(fd, false)
		defer sys.SetNonblock(fd, true)
	}

	_, err := w.file.Write(bytesBuf.Bytes())
	if err != nil {
		return err
	}

	w.oldBuf = buf
	return nil
}

// findWindow finds a window of lines around the selected line in a total
// number of height lines, that is at most max lines.
func findWindow(height, selected, max int) (low, high int) {
	if height <= max {
		// No need for windowing
		return 0, height
	}
	low = selected - max/2
	high = low + max
	switch {
	case low < 0:
		// Near top of the list, move the window down
		low = 0
		high = low + max
	case high > height:
		// Near bottom of the list, move the window down
		high = height
		low = high - max
	}
	return
}

func trimToWindow(s []string, selected, max int) ([]string, int) {
	low, high := findWindow(len(s), selected, max)
	return s[low:high], low
}

// refresh redraws the line editor. The dot is passed as an index into text;
// the corresponding position will be calculated.
func (w *Writer) refresh(es *editorState, fullRefresh bool) error {
	height, width := sys.GetWinsize(int(w.file.Fd()))
	er := &editorRenderer{es, height, nil}
	buf := render(er, width)
	return w.commitBuffer(er.bufNoti, buf, fullRefresh)
}
