package edit

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/util"
)

// Completion mode.

// Interface.

var _ = registerBuiltins(modeCompletion, map[string]func(*Editor){
	"smart-start":    complSmartStart,
	"start":          complStart,
	"up":             complUp,
	"up-cycle":       complUpCycle,
	"down":           complDown,
	"down-cycle":     complDownCycle,
	"left":           complLeft,
	"right":          complRight,
	"accept":         complAccept,
	"trigger-filter": complTriggerFilter,
	"default":        complDefault,
})

type completion struct {
	complSpec
	completer string

	filtering       bool
	filter          string
	filtered        []*candidate
	selected        int
	firstShown      int
	lastShownInFull int
	height          int
}

func (*completion) Binding(m map[string]vartypes.Variable, k ui.Key) eval.Fn {
	return getBinding(m[modeCompletion], k)
}

func (c *completion) needScrollbar() bool {
	return c.firstShown > 0 || c.lastShownInFull < len(c.filtered)-1
}

func (c *completion) ModeLine() ui.Renderer {
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

func complUpCycle(ed *Editor) {
	ed.completion.prev(true)
}

func complDownCycle(ed *Editor) {
	ed.completion.next(true)
}

// acceptCompletion accepts currently selected completion candidate.
func complAccept(ed *Editor) {
	c := ed.completion
	if 0 <= c.selected && c.selected < len(c.filtered) {
		ed.buffer, ed.dot = c.apply(ed.buffer, ed.dot)
	}
	ed.mode = &ed.insert
}

func complDefault(ed *Editor) {
	k := ed.lastKey
	c := &ed.completion
	if c.filtering && likeChar(k) {
		c.changeFilter(c.filter + string(k.Rune))
	} else if c.filtering && k == (ui.Key{ui.Backspace, 0}) {
		_, size := utf8.DecodeLastRuneInString(c.filter)
		if size > 0 {
			c.changeFilter(c.filter[:len(c.filter)-size])
		}
	} else {
		complAccept(ed)
		ed.setAction(reprocessKey)
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
	text := c.selectedCandidate().code
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

	completer, complSpec, err := complete(node, ed.evaler)

	if err != nil {
		ed.addTip("%v", err)
		// We don't show the full stack trace. To make debugging still possible,
		// we log it.
		if pprinter, ok := err.(util.Pprinter); ok {
			logger.Println("matcher error:")
			logger.Println(pprinter.Pprint(""))
		}
	} else if completer == "" {
		ed.addTip("unsupported completion :(")
		logger.Println("path to current leaf, leaf first")
		for n := node; n != nil; n = n.Parent() {
			logger.Printf("%T (%d-%d)", n, n.Begin(), n.End())
		}
	} else if len(complSpec.candidates) == 0 {
		ed.addTip("no candidate for %s", completer)
	} else {
		if acceptPrefix {
			// If there is a non-empty longest common prefix, insert it and
			// don't start completion mode.
			//
			// As a special case, when there is exactly one candidate, it is
			// immeidately accepted.
			prefix := complSpec.candidates[0].code
			for _, cand := range complSpec.candidates[1:] {
				prefix = commonPrefix(prefix, cand.code)
				if prefix == "" {
					break
				}
			}

			if prefix != "" && len(prefix) > complSpec.end-complSpec.begin {
				ed.buffer = ed.buffer[:complSpec.begin] + prefix + ed.buffer[complSpec.end:]
				ed.dot = complSpec.begin + len(prefix)

				return
			}
		}
		ed.completion = completion{
			completer: completer,
			complSpec: *complSpec,
			filtered:  complSpec.candidates,
		}
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
		w := util.Wcswidth(c.filtered[i].menu.Text)
		if width < w {
			width = w
		}
	}
	return width
}

func (c *completion) ListRender(width, maxHeight int) *ui.Buffer {
	b := ui.NewBuffer(width)
	cands := c.filtered
	if len(cands) == 0 {
		b.WriteString(util.TrimWcwidth("(no result)", width), "")
		return b
	}
	if maxHeight <= 1 || width <= 2 {
		b.WriteString(util.TrimWcwidth("(terminal too small)", width), "")
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

		col := ui.NewBuffer(totalColWidth)
		for j = i; j < i+height; j++ {
			if j > i {
				col.Newline()
			}
			if j >= len(cands) {
				// Write padding to make the listing a rectangle.
				col.WriteSpaces(totalColWidth, styleForCompletion.String())
			} else {
				col.WriteSpaces(completionColMarginLeft, styleForCompletion.String())
				s := ui.JoinStyles(styleForCompletion, cands[j].menu.Styles)
				if j == c.selected {
					s = append(s, styleForSelectedCompletion.String())
				}
				col.WriteString(util.ForceWcwidth(cands[j].menu.Text, colWidth), s.String())
				col.WriteSpaces(completionColMarginRight, styleForCompletion.String())
				if !trimmed {
					c.lastShownInFull = j
				}
			}
		}

		b.ExtendRight(col, 0)
		remainedWidth -= totalColWidth
		if remainedWidth <= completionColMarginTotal {
			break
		}
	}
	// When the listing is incomplete, always use up the entire width.
	if remainedWidth > 0 && c.needScrollbar() {
		col := ui.NewBuffer(remainedWidth)
		for i := 0; i < height; i++ {
			if i > 0 {
				col.Newline()
			}
			col.WriteSpaces(remainedWidth, styleForCompletion.String())
		}
		b.ExtendRight(col, 0)
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
		if strings.Contains(cand.menu.Text, f) {
			c.filtered = append(c.filtered, cand)
		}
	}
	if len(c.filtered) > 0 {
		c.selected = 0
	} else {
		c.selected = -1
	}
}
