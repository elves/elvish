package edit

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
	firstShown int
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

func complLeft(ed *Editor) {
	if c := ed.completion.selected - ed.completion.height; c >= 0 {
		ed.completion.selected = c
	}
}

func complRight(ed *Editor) {
	if c := ed.completion.selected + ed.completion.height; c < len(ed.completion.candidates) {
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
		// Fix .display.text
		for _, cand := range c.candidates {
			if cand.display.text == "" {
				cand.display.text = cand.text
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

// maxWidth finds the maximum wcwidth of display texts of candidates [lo, hi).
// hi may be larger than the number of candidates, in which case it is truncated
// to the number of candidates.
func (comp *completion) maxWidth(lo, hi int) int {
	if hi > len(comp.candidates) {
		hi = len(comp.candidates)
	}
	width := 0
	for i := lo; i < hi; i++ {
		w := WcWidths(comp.candidates[i].display.text)
		if width < w {
			width = w
		}
	}
	return width
}

func (comp *completion) List(width, maxHeight int) *buffer {
	b := newBuffer(width)
	cands := comp.candidates
	if len(cands) == 0 {
		b.writes(TrimWcWidth("(no result)", width), "")
		return b
	}

	comp.height = min(maxHeight, len(cands))

	// Determine the first column to show. We start with the column in which the
	// selected one is found, moving to the left until either the width is
	// exhausted, or the old value of firstShown has been hit.
	first := comp.selected / maxHeight * maxHeight
	w := comp.maxWidth(first, first+maxHeight)
	for ; first > comp.firstShown; first -= maxHeight {
		dw := comp.maxWidth(first-maxHeight, first) + completionListingColMargin
		if w+dw > width-2 {
			break
		}
		w += dw
	}
	comp.firstShown = first

	if first > 0 {
		// Draw a left arrow
		b.extendHorizontal(makeArrowColumn(maxHeight, "<"), 0)
	}

	var i, j int
	remainedWidth := width - 2
	margin := 0
	// Show the results in columns, until width is exceeded.
	for i = first; i < len(cands); i += maxHeight {
		if i > first {
			margin = completionListingColMargin
		}
		// Determine the width of the column (without the margin)
		colWidth := comp.maxWidth(i, min(i+maxHeight, len(cands)))
		if colWidth > remainedWidth-margin {
			colWidth = remainedWidth - margin
		}

		col := newBuffer(margin + colWidth)
		for j = i; j < i+maxHeight && j < len(cands); j++ {
			if j > i {
				col.newline()
			}
			col.writePadding(margin, "")
			style := cands[j].display.style
			if j == comp.selected {
				style += styleForSelected
			}
			col.writes(ForceWcWidth(cands[j].display.text, colWidth), style)
		}

		b.extendHorizontal(col, 1)
		remainedWidth -= colWidth + margin
		if remainedWidth <= completionListingColMargin {
			break
		}
	}
	if j < len(cands) {
		b.extendHorizontal(makeArrowColumn(maxHeight, ">"), 1)
	}
	return b
}

func makeArrowColumn(height int, s string) *buffer {
	b := newBuffer(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			b.newline()
		}
		if i == height/2 {
			b.writes(s, "")
		} else {
			b.writes(" ", "")
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
