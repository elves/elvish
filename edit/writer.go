package edit

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logWriterDetail = false

// writer renders the editor UI.
type writer struct {
	file   *os.File
	oldBuf *buffer
}

func newWriter(f *os.File) *writer {
	writer := &writer{file: f, oldBuf: &buffer{}}
	return writer
}

func (w *writer) resetOldBuf() {
	w.oldBuf = &buffer{}
}

// deltaPos calculates the escape sequence needed to move the cursor from one
// position to another. It use relative movements to move to the destination
// line and absolute movement to move to the destination column.
func deltaPos(from, to pos) []byte {
	buf := new(bytes.Buffer)
	if from.line < to.line {
		// move down
		fmt.Fprintf(buf, "\033[%dB", to.line-from.line)
	} else if from.line > to.line {
		// move up
		fmt.Fprintf(buf, "\033[%dA", from.line-to.line)
	}
	fmt.Fprintf(buf, "\033[%dG", to.col+1)
	return buf.Bytes()
}

// commitBuffer updates the terminal display to reflect current buffer.
// TODO Instead of erasing w.oldBuf entirely and then draw buf, compute a
// delta between w.oldBuf and buf
func (w *writer) commitBuffer(bufNoti, buf *buffer, fullRefresh bool) error {
	if buf.width != w.oldBuf.width && w.oldBuf.cells != nil {
		// Width change, force full refresh
		w.oldBuf.cells = nil
		fullRefresh = true
	}

	bytesBuf := new(bytes.Buffer)

	// Hide cursor.
	bytesBuf.WriteString("\033[?25l")

	// Rewind cursor
	if pLine := w.oldBuf.dot.line; pLine > 0 {
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

	writeCells := func(cs []cell) {
		for _, c := range cs {
			if c.width > 0 {
				switchStyle(c.style)
			}
			bytesBuf.WriteString(string(c.rune))
		}
	}

	if bufNoti != nil {
		if logWriterDetail {
			Logger.Printf("going to write %d lines of notifications", len(bufNoti.cells))
		}

		// Write notifications
		for _, line := range bufNoti.cells {
			writeCells(line)
			switchStyle("")
			bytesBuf.WriteString("\033[K\n")
		}
		// XXX Hacky.
		if len(w.oldBuf.cells) > 0 {
			w.oldBuf.cells = w.oldBuf.cells[1:]
		}
	}

	if logWriterDetail {
		Logger.Printf("going to write %d lines, oldBuf had %d", len(buf.cells), len(w.oldBuf.cells))
	}

	for i, line := range buf.cells {
		if i > 0 {
			bytesBuf.WriteString("\n")
		}
		var j int // First column where buf and oldBuf differ
		// No need to update current line
		if !fullRefresh && i < len(w.oldBuf.cells) {
			var eq bool
			if eq, j = compareRows(line, w.oldBuf.cells[i]); eq {
				continue
			}
		}
		// Move to the first differing column if necessary.
		firstCol := widthOfCells(line[:j])
		if firstCol != 0 {
			fmt.Fprintf(bytesBuf, "\033[%dG", firstCol+1)
		}
		// Erase the rest of the line if necessary.
		if !fullRefresh && i < len(w.oldBuf.cells) && j < len(w.oldBuf.cells[i]) {
			switchStyle("")
			bytesBuf.WriteString("\033[K")
		}
		writeCells(line[j:])
	}
	if len(w.oldBuf.cells) > len(buf.cells) && !fullRefresh {
		// If the old buffer is higher, erase old content.
		// Note that we cannot simply write \033[J, because if the cursor is
		// just over the last column -- which is precisely the case if we have a
		// rprompt, \033[J will also erase the last column.
		switchStyle("")
		bytesBuf.WriteString("\n\033[J\033[A")
	}
	switchStyle("")
	cursor := buf.cursor()
	bytesBuf.Write(deltaPos(cursor, buf.dot))

	// Show cursor.
	bytesBuf.WriteString("\033[?25h")

	if logWriterDetail {
		Logger.Printf("going to write %q", bytesBuf.String())
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
func (w *writer) refresh(es *editorState, fullRefresh bool) error {
	height, width := sys.GetWinsize(int(w.file.Fd()))
	mode := es.mode.Mode()

	var bufNoti, bufLine, bufMode, bufTips, bufListing, buf *buffer
	// butNoti
	if len(es.notifications) > 0 {
		bufNoti = newBuffer(width)
		bufNoti.writes(strings.Join(es.notifications, "\n"), "")
		es.notifications = nil
	}

	// bufLine
	b := newBuffer(width)
	bufLine = b

	b.newlineWhenFull = true

	b.writeStyleds(es.prompt)

	if b.line() == 0 && b.col*2 < b.width {
		b.indent = b.col
	}

	// i keeps track of number of bytes written.
	i := 0

	// nowAt is called at every rune boundary.
	nowAt := func(i int) {
		if mode == modeCompletion && i == es.completion.begin {
			c := es.completion.selectedCandidate()
			b.writes(c.text, styleForCompleted.String())
		}
		if i == es.dot {
			b.dot = b.cursor()
		}
	}
	nowAt(0)
tokens:
	for _, token := range es.tokens {
		for _, r := range token.Text {
			if mode == modeCompletion &&
				es.completion.begin <= i && i <= es.completion.end {
				// Do nothing. This part is replaced by the completion candidate.
			} else {
				b.write(r, joinStyles(styleForType[token.Type], token.MoreStyle).String())
			}
			i += utf8.RuneLen(r)

			nowAt(i)
			if mode == modeHistory && i == len(es.hist.prefix) {
				break tokens
			}
		}
	}

	if mode == modeHistory {
		// Put the rest of current history, position the cursor at the
		// end of the line, and finish writing
		h := es.hist
		b.writes(h.line[len(h.prefix):], styleForCompletedHistory.String())
		b.dot = b.cursor()
	}

	// Write rprompt
	if len(es.rprompt) > 0 {
		padding := b.width - b.col
		for _, s := range es.rprompt {
			padding -= util.Wcswidth(s.text)
		}
		if padding >= 1 {
			b.newlineWhenFull = false
			b.writePadding(padding, "")
			b.writeStyleds(es.rprompt)
		}
	}

	// bufMode
	bufMode = render(es.mode.ModeLine(), width)

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(es.tips) > 0 {
		bufTips = newBuffer(width)
		bufTips.writes(strings.Join(es.tips, "\n"), styleForTip.String())
	}

	hListing := 0
	// Trim lines and determine the maximum height for bufListing
	// TODO come up with a UI to tell the user that something is not shown.
	switch {
	case height >= lines(bufNoti, bufLine, bufMode, bufTips):
		hListing = height - lines(bufLine, bufMode, bufTips)
	case height >= lines(bufNoti, bufLine, bufTips):
		bufMode = nil
	case height >= lines(bufNoti, bufLine):
		bufMode = nil
		if bufTips != nil {
			bufTips.trimToLines(0, height-lines(bufNoti, bufLine))
		}
	case height >= lines(bufLine):
		bufTips, bufMode = nil, nil
		if bufNoti != nil {
			n := len(bufNoti.cells)
			bufNoti.trimToLines(n-(height-lines(bufLine)), n)
		}
	case height >= 1:
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine+1-height, dotLine+1)
	default:
		// Broken terminal. Still try to render one line of bufLine.
		bufNoti, bufTips, bufMode = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine, dotLine+1)
	}

	// bufListing.
	if hListing > 0 {
		if lister, ok := es.mode.(OldLister); ok {
			bufListing = lister.List(width, hListing)
		} else if lister, ok := es.mode.(Lister); ok {
			bufListing = render(lister.List(hListing), width)
		}
		// XXX When in completion mode, we re-render the mode line, since the
		// scrollbar in the mode line depends on completion.lastShown which is
		// only known after the listing has been rendered. Since rendering the
		// scrollbar never adds additional lines to bufMode, we may do this
		// without recalculating the layout.
		if mode == modeCompletion {
			bufMode = render(es.mode.ModeLine(), width)
		}
	}

	if logWriterDetail {
		Logger.Printf("bufLine %d, bufMode %d, bufTips %d, bufListing %d",
			lines(bufLine), lines(bufMode), lines(bufTips), lines(bufListing))
	}

	// Combine buffers (reusing bufLine)
	buf = bufLine
	buf.extend(bufMode, mode == modeLocation || mode == modeHistoryListing ||
		(mode == modeCompletion && es.completion.filtering) ||
		(mode == modeNavigation && es.navigation.filtering))
	buf.extend(bufTips, false)
	buf.extend(bufListing, false)

	return w.commitBuffer(bufNoti, buf, fullRefresh)
}
