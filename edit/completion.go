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

	filtering       bool
	filter          string
	candidates      []*candidate
	selected        int
	firstShown      int
	lastShownInFull int
	height          int
}

func (*completion) Mode() ModeType {
	return modeCompletion
}

func (c *completion) needScrollbar() bool {
	return c.firstShown > 0 || c.lastShownInFull < len(c.candidates)-1
}

func (c *completion) ModeLine() renderer {
	ml := modeLineRenderer{fmt.Sprintf(" COMPLETING %s ", c.completer), c.filter}
	if !c.needScrollbar() {
		return ml
	}
	return modeLineWithScrollBarRenderer{ml,
		len(c.candidates), c.firstShown, c.lastShownInFull + 1}
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
	for _, item := range completers {
		compl, err := item.completer(node, ed.evaler)
		if compl != nil {
			c.completer = item.name
			c.begin, c.end, c.all = compl.begin, compl.end, compl.cands
			c.candidates = c.all
			break
		} else if err != nil && err != errCompletionUnapplicable {
			ed.Notify("%v", err)
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

const (
	completionColMarginLeft  = 1
	completionColMarginRight = 1
	completionColMarginTotal = completionColMarginLeft + completionColMarginRight
)

// maxWidth finds the maximum wcwidth of display texts of candidates [lo, hi).
// hi may be larger than the number of candidates, in which case it is truncated
// to the number of candidates.
func (comp *completion) maxWidth(lo, hi int) int {
	if hi > len(comp.candidates) {
		hi = len(comp.candidates)
	}
	width := 0
	for i := lo; i < hi; i++ {
		w := util.Wcswidth(comp.candidates[i].display.text)
		if width < w {
			width = w
		}
	}
	return width
}

func (comp *completion) ListRender(width, maxHeight int) *buffer {
	b := newBuffer(width)
	cands := comp.candidates
	if len(cands) == 0 {
		b.writes(util.TrimWcwidth("(no result)", width), "")
		return b
	}
	if maxHeight <= 1 || width <= 2 {
		b.writes(util.TrimWcwidth("(terminal too small)", width), "")
		return b
	}

	// Reserve the the rightmost row as margins.
	width -= 1

	// Determine comp.height and comp.firstShown.
	// First determine whether all candidates can be fit in the screen,
	// assuming that they are all of maximum width. If that is the case, we use
	// the computed height as the height for the listing, and the first
	// candidate to show is 0. Otherwise, we use min(height, len(cands)) as the
	// height and find the first candidate to show.
	perLine := max(1, width/(comp.maxWidth(0, len(cands))+completionColMarginTotal))
	heightBound := util.CeilDiv(len(cands), perLine)
	first := 0
	height := 0
	if heightBound < maxHeight {
		height = heightBound
	} else {
		height = min(maxHeight, len(cands))
		// Determine the first column to show. We start with the column in which the
		// selected one is found, moving to the left until either the width is
		// exhausted, or the old value of firstShown has been hit.
		first = comp.selected / height * height
		w := comp.maxWidth(first, first+height) + completionColMarginTotal
		for ; first > comp.firstShown; first -= height {
			dw := comp.maxWidth(first-height, first) + completionColMarginTotal
			if w+dw > width {
				break
			}
			w += dw
		}
	}
	comp.height = height
	comp.firstShown = first

	var i, j int
	remainedWidth := width
	trimmed := false
	// Show the results in columns, until width is exceeded.
	for i = first; i < len(cands); i += height {
		// Determine the width of the column (without the margin)
		colWidth := comp.maxWidth(i, min(i+height, len(cands)))
		totalColWidth := colWidth + completionColMarginTotal
		if totalColWidth > remainedWidth {
			totalColWidth = remainedWidth
			colWidth = totalColWidth - completionColMarginTotal
			trimmed = true
		}

		col := newBuffer(totalColWidth)
		for j = i; j < i+height; j++ {
			if j > i {
				col.newline()
			}
			if j >= len(cands) {
				// Write padding to make the listing a rectangle.
				col.writePadding(totalColWidth, styleForCompletion.String())
			} else {
				col.writePadding(completionColMarginLeft, styleForCompletion.String())
				s := joinStyles(styleForCompletion, cands[j].display.styles)
				if j == comp.selected {
					s = append(s, styleForSelectedCompletion.String())
				}
				col.writes(util.ForceWcwidth(cands[j].display.text, colWidth), s.String())
				col.writePadding(completionColMarginRight, styleForCompletion.String())
				if !trimmed {
					comp.lastShownInFull = j
				}
			}
		}

		b.extendHorizontal(col, 0)
		remainedWidth -= totalColWidth
		if remainedWidth <= completionColMarginTotal {
			break
		}
	}
	// When the listing is incomplete, always use up the entire width.
	if remainedWidth > 0 && comp.needScrollbar() {
		col := newBuffer(remainedWidth)
		for i := 0; i < height; i++ {
			if i > 0 {
				col.newline()
			}
			col.writePadding(remainedWidth, styleForCompletion.String())
		}
		b.extendHorizontal(col, 0)
		remainedWidth = 0
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
