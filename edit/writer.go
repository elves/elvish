package edit

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/elves/elvish/sys"
)

const (
	completionListingColMargin          int = 2
	navigationListingColMargin              = 1
	navigationListingColPadding             = 1
	navigationListingMinWidthForPadding     = 5
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

var invalidPos = pos{-1, -1}

func lineWidth(cs []cell) int {
	w := 0
	for _, c := range cs {
		w += int(c.width)
	}
	return w
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

func makeSpacing(n int) []cell {
	s := make([]cell, n)
	for i := 0; i < n; i++ {
		s[i].rune = ' '
		s[i].width = 1
	}
	return s
}

// extendHorizontal extends b horizontally, appending each line in b2 to b,
// preceeded by m margin. If b2 has more lines than b, the last len(b2) -
// len(b) lines are first filled with paddings of width w.
func (b *buffer) extendHorizontal(b2 *buffer, w, m int) {
	i := 0
	margin := makeSpacing(m)
	for ; i < len(b.cells) && i < len(b2.cells); i++ {
		if w0 := lineWidth(b.cells[i]); w0 < w {
			b.cells[i] = append(b.cells[i], makeSpacing(w-w0)...)
		}
		b.cells[i] = append(append(b.cells[i], margin...), b2.cells[i]...)
	}
	padding := makeSpacing(w + m)
	for ; i < len(b2.cells); i++ {
		row := make([]cell, 0, w+m+len(b2.cells[i]))
		row = append(append(row, padding...), b2.cells[i]...)
		b.cells = append(b.cells, row)
	}
}

// write appends a single rune to a buffer.
func (b *buffer) write(r rune, attr string) {
	if r == '\n' {
		b.newline()
		return
	} else if !unicode.IsPrint(r) {
		// BUG(xiaq): buffer.write drops unprintable runes silently
		return
	}
	wd := WcWidth(r)
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
		fmt.Fprintf(buf, "\033[%dB", to.line-from.line)
	} else if from.line > to.line {
		// move up
		fmt.Fprintf(buf, "\033[%dA", from.line-to.line)
	}
	fmt.Fprintf(buf, "\033[%dG", to.col+1)
	return buf.Bytes()
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

// commitBuffer updates the terminal display to reflect current buffer.
// TODO Instead of erasing w.oldBuf entirely and then draw buf, compute a
// delta between w.oldBuf and buf
func (w *writer) commitBuffer(buf *buffer) error {
	var fullRefresh bool
	if buf.width != w.oldBuf.width && w.oldBuf.cells != nil {
		// Width change, force full refresh
		w.oldBuf.cells = nil
		fullRefresh = true
	}

	bytesBuf := new(bytes.Buffer)

	// Rewind cursor
	if pLine := w.oldBuf.dot.line; pLine > 0 {
		fmt.Fprintf(bytesBuf, "\033[%dA", pLine)
	}
	bytesBuf.WriteString("\r")

	attr := ""
	for i, line := range buf.cells {
		if i > 0 {
			bytesBuf.WriteString("\n")
		}
		var j int // First column where buf and oldBuf differ
		// No need to update current line
		if i < len(w.oldBuf.cells) {
			var eq bool
			if eq, j = compareRows(line, w.oldBuf.cells[i]); eq {
				continue
			}
		}
		// Move to the first differing column and erase the rest of line
		fmt.Fprintf(bytesBuf, "\033[%dG\033[K", j+1)
		for _, c := range line[j:] {
			if c.width > 0 && c.attr != attr {
				fmt.Fprintf(bytesBuf, "\033[m\033[%sm", c.attr)
				attr = c.attr
			}
			bytesBuf.WriteString(string(c.rune))
		}
	}
	// If the old buffer is higher, erase old content
	if len(w.oldBuf.cells) > len(buf.cells) || fullRefresh {
		bytesBuf.WriteString("\n\033[J\033[A")
	}
	if attr != "" {
		bytesBuf.WriteString("\033[m")
	}
	cursor := buf.cursor()
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

func renderNavColumn(nc *navColumn, w, h int) *buffer {
	b := newBuffer(w)
	low, high := findWindow(len(nc.names), nc.selected, h)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		text := nc.names[i]
		attr := nc.attrs[i]
		if i == nc.selected {
			attr += attrForSelectedFile
		}
		if w >= navigationListingMinWidthForPadding {
			padding := navigationListingColPadding
			b.writePadding(padding, attr)
			b.writes(ForceWcWidth(text, w-2), attr)
			b.writePadding(padding, attr)
		} else {
			b.writes(ForceWcWidth(text, w), attr)
		}
	}
	return b
}

// refresh redraws the line editor. The dot is passed as an index into text;
// the corresponding position will be calculated.
func (w *writer) refresh(es *editorState) error {
	winsize := sys.GetWinsize(int(w.file.Fd()))
	width, height := int(winsize.Col), int(winsize.Row)

	var bufLine, bufMode, bufTips, bufListing, buf *buffer
	// bufLine
	b := newBuffer(width)
	bufLine = b

	b.newlineWhenFull = true

	b.writes(es.prompt, attrForPrompt)

	if b.line() == 0 && b.col*2 < b.width {
		b.indent = b.col
	}

	// i keeps track of number of bytes written.
	i := 0
	if es.dot == 0 {
		b.dot = b.cursor()
	}

	comp := es.completion
	hasComp := comp != nil && comp.current != -1

	nowAt := func(i int) {
		if hasComp && comp.start == i {
			// Put the current completion candidate.
			for _, part := range comp.candidates[comp.current].parts {
				attr := attrForType[comp.typ]
				if part.completed {
					attr += attrForCompleted
				}
				b.writes(part.text, attr)
			}
		}
		if es.dot == i {
			b.dot = b.cursor()
		}
	}
	nowAt(0)
tokens:
	for _, token := range es.tokens {
		for _, r := range token.Text {
			if hasComp && comp.start <= i && i < comp.end {
				// Silence the part that is being completed
			} else {
				b.write(r, attrForType[token.Type]+token.MoreStyle)
			}
			i += utf8.RuneLen(r)

			nowAt(i)
			if es.mode == modeHistory && i == len(es.history.prefix) {
				break tokens
			}
		}
	}

	if es.mode == modeHistory {
		// Put the rest of current history, position the cursor at the
		// end of the line, and finish writing
		h := es.history
		b.writes(h.line[len(h.prefix):], attrForCompletedHistory)
		b.dot = b.cursor()
	}

	// Write rprompt
	padding := b.width - b.col - WcWidths(es.rprompt)
	if padding >= 1 {
		b.newlineWhenFull = false
		b.writePadding(padding, "")
		b.writes(es.rprompt, attrForRprompt)
	}

	// bufMode
	if es.mode != modeInsert {
		b := newBuffer(width)
		bufMode = b
		text := ""
		switch es.mode {
		case modeCommand:
			text = "COMMAND"
		case modeCompletion:
			text = fmt.Sprintf("COMPLETING %s", comp.completer)
		case modeNavigation:
			text = "NAVIGATING"
		case modeHistory:
			text = fmt.Sprintf("HISTORY #%d", es.history.current)
		}
		b.writes(TrimWcWidth(" "+text+" ", width), attrForMode)
	}

	// bufTips
	// TODO tips is assumed to contain no newlines.
	if len(es.tips) > 0 {
		b := newBuffer(width)
		bufTips = b
		b.writes(TrimWcWidth(strings.Join(es.tips, ", "), width), attrForTip)
	}

	hListing := 0
	// Trim lines and determine the maximum height for bufListing
	switch {
	case height >= lines(bufLine, bufMode, bufTips):
		hListing = height - lines(bufLine, bufMode, bufTips)
	case height >= lines(bufLine, bufTips):
		bufMode, bufListing = nil, nil
	case height >= lines(bufLine):
		bufTips, bufMode, bufListing = nil, nil, nil
	case height >= 1:
		bufTips, bufMode, bufListing = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine+1-height, dotLine+1)
	default:
		// Broken terminal. Still try to render one line of bufLine.
		bufTips, bufMode, bufListing = nil, nil, nil
		dotLine := bufLine.dot.line
		bufLine.trimToLines(dotLine, dotLine+1)
	}

	// Render bufListing under the maximum height constraint
	nav := es.navigation
	if hListing > 0 && comp != nil || nav != nil {
		b := newBuffer(width)
		bufListing = b
		// Completion listing
		if comp != nil {
			// Layout candidates in multiple columns
			cands := comp.candidates

			// First decide the shape (# of rows and columns)
			colWidth := 0
			margin := completionListingColMargin
			for _, cand := range cands {
				width := WcWidths(cand.text)
				if colWidth < width {
					colWidth = width
				}
			}

			cols := (b.width + margin) / (colWidth + margin)
			if cols == 0 {
				cols = 1
			}
			lines := CeilDiv(len(cands), cols)
			es.completionLines = lines

			// Determine the window to show.
			low, high := findWindow(lines, comp.current%lines, hListing)
			for i := low; i < high; i++ {
				if i > low {
					b.newline()
				}
				for j := 0; j < cols; j++ {
					k := j*lines + i
					if k >= len(cands) {
						continue
					}
					attr := cands[k].attr
					if k == comp.current {
						attr += attrForCurrentCompletion
					}
					text := cands[k].text
					b.writes(ForceWcWidth(text, colWidth), attr)
					b.writePadding(margin, "")
				}
			}
		}

		// Navigation listing
		if nav != nil {
			margin := navigationListingColMargin
			var ratioParent, ratioCurrent, ratioPreview int
			if nav.dirPreview != nil {
				ratioParent = 15
				ratioCurrent = 40
				ratioPreview = 45
			} else {
				ratioParent = 15
				ratioCurrent = 75
				// Leave some space at the right side
			}

			w := width - margin*2

			wParent := w * ratioParent / 100
			wCurrent := w * ratioCurrent / 100
			wPreview := w * ratioPreview / 100

			b := renderNavColumn(nav.parent, wParent, hListing)
			bufListing = b

			bCurrent := renderNavColumn(nav.current, wCurrent, hListing)
			b.extendHorizontal(bCurrent, wParent, margin)

			if wPreview > 0 {
				bPreview := renderNavColumn(nav.dirPreview, wPreview, hListing)
				b.extendHorizontal(bPreview, wParent+wCurrent+margin, margin)
			}
		}
	}

	// Combine buffers (reusing bufLine)
	buf = bufLine
	buf.extend(bufMode)
	buf.extend(bufTips)
	buf.extend(bufListing)

	return w.commitBuffer(buf)
}
