package edit

import (
	"fmt"
	"unicode/utf8"

	"github.com/elves/elvish/util"
)

// Completion subsystem.

// Interface.

type completion struct {
	completer  string
	begin, end int
	candidates []*candidate
	selected   int
	lines      int
}

func (*completion) Mode() ModeType {
	return modeCompletion
}

func (c *completion) ModeLine(width int) *buffer {
	return makeModeLine(fmt.Sprintf("COMPLETING %s", c.completer), width)
}

func startCompletion(ed *Editor) {
	startCompletionInner(ed, false)
}

func completePrefixOrStartCompletion(ed *Editor) {
	startCompletionInner(ed, true)
}

func selectCandUp(ed *Editor) {
	ed.completion.prev(false)
}

func selectCandDown(ed *Editor) {
	ed.completion.next(false)
}

func selectCandLeft(ed *Editor) {
	if c := ed.completion.selected - ed.completion.lines; c >= 0 {
		ed.completion.selected = c
	}
}

func selectCandRight(ed *Editor) {
	if c := ed.completion.selected + ed.completion.lines; c < len(ed.completion.candidates) {
		ed.completion.selected = c
	}
}

func cycleCandRight(ed *Editor) {
	ed.completion.next(true)
}

// acceptCompletion accepts currently selected completion candidate.
func acceptCompletion(ed *Editor) {
	c := ed.completion
	if 0 <= c.selected && c.selected < len(c.candidates) {
		ed.line, ed.dot = c.apply(ed.line, ed.dot)
	}
	ed.mode = &ed.insert
}

func defaultCompletion(ed *Editor) {
	acceptCompletion(ed)
	ed.nextAction = action{typ: reprocessKey}
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
			c.begin, c.end, c.candidates = begin, end, candidates
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

	cols := (b.width + margin) / (colWidth + margin)
	if cols == 0 {
		cols = 1
	}
	lines := util.CeilDiv(len(cands), cols)
	comp.lines = lines

	// Determine the window to show.
	low, high := findWindow(lines, comp.selected%lines, maxHeight)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		for j := 0; j < cols; j++ {
			k := j*lines + i
			if k >= len(cands) {
				break
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
	}
	return b
}
