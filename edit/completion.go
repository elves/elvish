package edit

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/uitypes"
	"github.com/elves/elvish/util"
)

// Completion subsystem.

// Interface.

var _ = registerBuiltins("compl", map[string]func(*Editor){
	"smart-start":    complSmartStart,
	"start":          complStart,
	"up":             complUp,
	"down":           complDown,
	"down-cycle":     complDownCycle,
	"left":           complLeft,
	"right":          complRight,
	"accept":         complAccept,
	"trigger-filter": complTriggerFilter,
	"default":        complDefault,
})

func init() {
	registerBindings(modeCompletion, "compl", map[uitypes.Key]string{
		{uitypes.Up, 0}:     "up",
		{uitypes.Down, 0}:   "down",
		{uitypes.Tab, 0}:    "down-cycle",
		{uitypes.Left, 0}:   "left",
		{uitypes.Right, 0}:  "right",
		{uitypes.Enter, 0}:  "accept",
		{'F', uitypes.Ctrl}: "trigger-filter",
		{'[', uitypes.Ctrl}: "insert:start",
		uitypes.Default:     "default",
	})
}

type completion struct {
	compl
	completer string

	filtering       bool
	filter          string
	filtered        []*candidate
	selected        int
	firstShown      int
	lastShownInFull int
	height          int
}

func (*completion) Mode() ModeType {
	return modeCompletion
}

func (c *completion) needScrollbar() bool {
	return c.firstShown > 0 || c.lastShownInFull < len(c.filtered)-1
}

func (c *completion) ModeLine() renderer {
	ml := modeLineRenderer{fmt.Sprintf(" COMPLETING %s ", c.completer), c.filter}
	if !c.needScrollbar() {
		return ml
	}
	return modeLineWithScrollBarRenderer{ml,
		len(c.filtered), c.firstShown, c.lastShownInFull + 1}
}

func (c *completion) CursorOnModeLine() bool {
	return c.filtering
}

func complStart(ed *Editor) {
	startCompletionInner(ed, false)
}

func complSmartStart(ed *Editor) {
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
	if c := ed.completion.selected + ed.completion.height; c < len(ed.completion.filtered) {
		ed.completion.selected = c
	}
}

func complDownCycle(ed *Editor) {
	ed.completion.next(true)
}

// acceptCompletion accepts currently selected completion candidate.
func complAccept(ed *Editor) {
	c := ed.completion
	if 0 <= c.selected && c.selected < len(c.filtered) {
		ed.line, ed.dot = c.apply(ed.line, ed.dot)
	}
	ed.mode = &ed.insert
}

func complDefault(ed *Editor) {
	k := ed.lastKey
	c := &ed.completion
	if c.filtering && likeChar(k) {
		c.changeFilter(c.filter + string(k.Rune))
	} else if c.filtering && k == (uitypes.Key{uitypes.Backspace, 0}) {
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

func (c *completion) selectedCandidate() *candidate {
	if c.selected == -1 {
		return &candidate{}
	}
	return c.filtered[c.selected]
}

// apply returns the line and dot after applying a candidate.
func (c *completion) apply(line string, dot int) (string, int) {
	text := c.selectedCandidate().text
	return line[:c.begin] + text + line[c.end:], c.begin + len(text)
}

func (c *completion) prev(cycle bool) {
	c.selected--
	if c.selected == -1 {
		if cycle {
			c.selected = len(c.filtered) - 1
		} else {
			c.selected++
		}
	}
}

func (c *completion) next(cycle bool) {
	c.selected++
	if c.selected == len(c.filtered) {
		if cycle {
			c.selected = 0
		} else {
			c.selected--
		}
	}
}

func startCompletionInner(ed *Editor, acceptPrefix bool) {
	node := findLeafNode(ed.chunk, ed.dot)
	if node == nil {
		return
	}

	c := &completion{}
	shownError := false
	for _, item := range completers {
		compl, err := item.completer(node, ed.evaler)
		if compl != nil {
			c.completer = item.name
			c.compl = *compl
			c.filtered = c.candidates
			break
		} else if err != nil && err != errCompletionUnapplicable {
			ed.addTip("%v", err)
			shownError = true
			break
		}
	}

	if c.completer == "" {
		if !shownError {
			ed.addTip("unsupported completion :(")
		}
		logger.Println("path to current leaf, leaf first")
		for n := node; n != nil; n = n.Parent() {
			logger.Printf("%T (%d-%d)", n, n.Begin(), n.End())
		}
	} else if len(c.filtered) == 0 {
		ed.addTip("no candidate for %s", c.completer)
	} else {
		if acceptPrefix {
			// If there is a non-empty longest common prefix, insert it and
			// don't start completion mode.
			//
			// As a special case, when there is exactly one candidate, it is
			// immeidately accepted.
			prefix := c.filtered[0].text
			for _, cand := range c.filtered[1:] {
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
func (c *completion) maxWidth(lo, hi int) int {
	if hi > len(c.filtered) {
		hi = len(c.filtered)
	}
	width := 0
	for i := lo; i < hi; i++ {
		w := util.Wcswidth(c.filtered[i].display.text)
		if width < w {
			width = w
		}
	}
	return width
}

func (c *completion) ListRender(width, maxHeight int) *buffer {
	b := newBuffer(width)
	cands := c.filtered
	if len(cands) == 0 {
		b.writes(util.TrimWcwidth("(no result)", width), "")
		return b
	}
	if maxHeight <= 1 || width <= 2 {
		b.writes(util.TrimWcwidth("(terminal too small)", width), "")
		return b
	}

	// Reserve the the rightmost row as margins.
	width--

	// Determine comp.height and comp.firstShown.
	// First determine whether all candidates can be fit in the screen,
	// assuming that they are all of maximum width. If that is the case, we use
	// the computed height as the height for the listing, and the first
	// candidate to show is 0. Otherwise, we use min(height, len(cands)) as the
	// height and find the first candidate to show.
	perLine := max(1, width/(c.maxWidth(0, len(cands))+completionColMarginTotal))
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
		first = c.selected / height * height
		w := c.maxWidth(first, first+height) + completionColMarginTotal
		for ; first > c.firstShown; first -= height {
			dw := c.maxWidth(first-height, first) + completionColMarginTotal
			if w+dw > width {
				break
			}
			w += dw
		}
	}
	c.height = height
	c.firstShown = first

	var i, j int
	remainedWidth := width
	trimmed := false
	// Show the results in columns, until width is exceeded.
	for i = first; i < len(cands); i += height {
		// Determine the width of the column (without the margin)
		colWidth := c.maxWidth(i, min(i+height, len(cands)))
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
				if j == c.selected {
					s = append(s, styleForSelectedCompletion.String())
				}
				col.writes(util.ForceWcwidth(cands[j].display.text, colWidth), s.String())
				col.writePadding(completionColMarginRight, styleForCompletion.String())
				if !trimmed {
					c.lastShownInFull = j
				}
			}
		}

		b.extendRight(col, 0)
		remainedWidth -= totalColWidth
		if remainedWidth <= completionColMarginTotal {
			break
		}
	}
	// When the listing is incomplete, always use up the entire width.
	if remainedWidth > 0 && c.needScrollbar() {
		col := newBuffer(remainedWidth)
		for i := 0; i < height; i++ {
			if i > 0 {
				col.newline()
			}
			col.writePadding(remainedWidth, styleForCompletion.String())
		}
		b.extendRight(col, 0)
		remainedWidth = 0
	}
	return b
}

func (c *completion) changeFilter(f string) {
	c.filter = f
	if f == "" {
		c.filtered = c.candidates
		return
	}
	c.filtered = nil
	for _, cand := range c.candidates {
		if strings.Contains(cand.display.text, f) {
			c.filtered = append(c.filtered, cand)
		}
	}
	if len(c.filtered) > 0 {
		c.selected = 0
	} else {
		c.selected = -1
	}
}
