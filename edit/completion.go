package edit

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/util"
)

// Completion subsystem.

// Interface.

type completion struct {
	completer  string
	begin, end int
	all        []*candidate

	filtering  bool
	filter     string
	candidates []*candidate
	selected   int
	lines      int
	height     int
}

func (*completion) Mode() ModeType {
	return modeCompletion
}

func (c *completion) ModeLine(width int) *buffer {
	title := fmt.Sprintf("COMPLETING %s", c.completer)
	// XXX Copied from listing.ModeLine.
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(TrimWcWidth(title, width), styleForMode)
	b.writes(" ", "")
	b.writes(c.filter, styleForFilter)
	b.dot = b.cursor()
	return b
}

func startCompl(ed *Editor) {
	startCompletionInner(ed, false)
}

func complPrefixOrStartCompl(ed *Editor) {
	startCompletionInner(ed, true)
}

func complUp(ed *Editor) {
	ed.completion.prev(false)
}

func complDown(ed *Editor) {
	ed.completion.next(false)
}

func complPageUp(ed *Editor) {
	ed.completion.pageUp()
}

func complPageDown(ed *Editor) {
	ed.completion.pageDown()
}

func complLeft(ed *Editor) {
	if c := ed.completion.selected - ed.completion.lines; c >= 0 {
		ed.completion.selected = c
	}
}

func complRight(ed *Editor) {
	if c := ed.completion.selected + ed.completion.lines; c < len(ed.completion.candidates) {
		ed.completion.selected = c
	}
}

func complDownCycle(ed *Editor) {
	ed.completion.next(true)
}

// acceptCompletion accepts currently selected completion candidate.
func complAccept(ed *Editor) {
	c := ed.completion
	if 0 <= c.selected && c.selected < len(c.candidates) {
		ed.line, ed.dot = c.apply(ed.line, ed.dot)
	}
	ed.mode = &ed.insert
}

func complDefault(ed *Editor) {
	k := ed.lastKey
	c := &ed.completion
	if c.filtering && likeChar(k) {
		c.changeFilter(c.filter + string(k.Rune))
	} else if c.filtering && k == (Key{Backspace, 0}) {
		_, size := utf8.DecodeLastRuneInString(c.filter)
		if size > 0 {
			c.changeFilter(c.filter[:len(c.filter)-size])
		}
	} else {
		complAccept(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func complTriggerFilter(ed *Editor) {
	c := &ed.completion
	if c.filtering {
		c.filtering = false
		c.changeFilter("")
	} else {
		c.filtering = true
	}
}

// Implementation.

type styled struct {
	text  string
	style string
}

type candidate struct {
	text    string
	display styled
	suffix  string
}

func (comp *completion) selectedCandidate() *candidate {
	if comp.selected == -1 {
		return &candidate{}
	}
	return comp.candidates[comp.selected]
}

// apply returns the line and dot after applying a candidate.
func (comp *completion) apply(line string, dot int) (string, int) {
	text := comp.selectedCandidate().text
	return line[:comp.begin] + text + line[comp.end:], comp.begin + len(text)
}

func (c *completion) prev(cycle bool) {
	c.selected--
	if c.selected == -1 {
		if cycle {
			c.selected = len(c.candidates) - 1
		} else {
			c.selected++
		}
	}
}

func (c *completion) next(cycle bool) {
	c.selected++
	if c.selected == len(c.candidates) {
		if cycle {
			c.selected = 0
		} else {
			c.selected--
		}
	}
}

func (c *completion) pageUp() {
	line, col := c.currentCoord()
	line -= c.height
	if line < 0 {
		line = 0
	}
	c.selected = line + col*c.lines
}

func (c *completion) pageDown() {
	line, col := c.currentCoord()
	line += c.height
	if line >= c.lines {
		line = c.lines - 1
	}
	c.selected = line + col*c.lines
	if c.selected >= len(c.candidates) {
		c.selected = len(c.candidates) - 1
	}
}

func (c *completion) currentCoord() (int, int) {
	return c.selected % c.lines, c.selected / c.lines
}

func startCompletionInner(ed *Editor, acceptPrefix bool) {
	token := tokenAtDot(ed)
	node := token.Node
	if node == nil {
		return
	}

	c := &completion{begin: -1}
	for _, compl := range completers {
		begin, end, candidates := compl.completer(node, ed)
		if begin >= 0 {
			c.completer = compl.name
			c.begin, c.end, c.all = begin, end, candidates
			c.candidates = c.all
			break
		}
	}

	if c.begin < 0 {
		ed.addTip("unsupported completion :(")
	} else if len(c.candidates) == 0 {
		ed.addTip("no candidate for %s", c.completer)
	} else {
		if acceptPrefix {
			// If there is a non-empty longest common prefix, insert it and
			// don't start completion mode.
			//
			// As a special case, when there is exactly one candidate, it is
			// immeidately accepted.
			prefix := c.candidates[0].text
			for _, cand := range c.candidates[1:] {
				prefix = commonPrefix(prefix, cand.text)
				if prefix == "" {
					break
				}
			}
			if prefix != "" && prefix != ed.line[c.begin:c.end] {
				ed.line = ed.line[:c.begin] + prefix + ed.line[c.end:]
				ed.dot = c.begin + len(prefix)
				return
			}
		}
		ed.completion = *c
		ed.mode = &ed.completion
	}
}

func tokenAtDot(ed *Editor) Token {
	if len(ed.tokens) == 0 || ed.dot > len(ed.line) {
		return Token{}
	}
	if ed.dot == len(ed.line) {
		return ed.tokens[len(ed.tokens)-1]
	}
	for _, token := range ed.tokens {
		if ed.dot < token.Node.End() {
			return token
		}
	}
	return Token{}
}

// commonPrefix returns the longest common prefix of two strings.
func commonPrefix(s, t string) string {
	for i, r := range s {
		if i >= len(t) {
			return s[:i]
		}
		r2, _ := utf8.DecodeRuneInString(t[i:])
		if r2 != r {
			return s[:i]
		}
	}
	return s
}

const completionListingColMargin = 2

func (comp *completion) List(width, maxHeight int) *buffer {
	b := newBuffer(width)
	// Layout candidates in multiple columns
	cands := comp.candidates

	if len(cands) == 0 {
		b.writes(TrimWcWidth("(no result)", width), "")
		return b
	}

	// First decide the shape (# of rows and columns)
	colWidth := 0
	margin := completionListingColMargin
	for _, cand := range cands {
		// XXX we also patch cand.display.text here.
		if cand.display.text == "" {
			cand.display.text = cand.text
		}
		width := WcWidths(cand.display.text)
		if colWidth < width {
			colWidth = width
		}
	}

	findShape := func(width int) (int, int, int) {
		cols := (width + margin) / (colWidth + margin)
		if cols == 0 {
			cols = 1
		}
		lines := util.CeilDiv(len(cands), cols)
		return cols, lines, width - colWidth*cols - margin*(cols-1)
	}
	cols, lines, rightspare := findShape(width)
	showScrollbar := lines > maxHeight && width > 1
	if showScrollbar && rightspare == 0 {
		cols, lines, rightspare = findShape(width - 1)
		rightspare++
	}
	comp.lines = lines

	// Determine the window to show.
	low, high := findWindow(lines, comp.selected%lines, maxHeight)
	comp.height = high - low
	var scrollLow, scrollHigh int
	if showScrollbar {
		scrollLow, scrollHigh = findScrollInterval(lines, low, high)
	}
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		for j := 0; j < cols; j++ {
			k := j*lines + i
			if k >= len(cands) {
				if j > 0 {
					b.writePadding(margin, "")
				}
				b.writePadding(colWidth, "")
				continue
			}
			style := cands[k].display.style
			if k == comp.selected {
				style += styleForSelected
			}
			text := cands[k].display.text
			if j > 0 {
				b.writePadding(margin, "")
			}
			b.writes(ForceWcWidth(text, colWidth), style)
		}

		if showScrollbar {
			bar := "│"
			if scrollLow <= i && i < scrollHigh {
				bar = "▉"
			}
			b.writePadding(rightspare-1, "")
			b.writes(bar, styleForScrollBar)
		}
	}
	return b
}

func (c *completion) changeFilter(f string) {
	c.filter = f
	if f == "" {
		c.candidates = c.all
		return
	}
	c.candidates = nil
	for _, cand := range c.all {
		if strings.Contains(cand.display.text, f) {
			c.candidates = append(c.candidates, cand)
		}
	}
	if len(c.candidates) > 0 {
		c.selected = 0
	} else {
		c.selected = -1
	}
}
