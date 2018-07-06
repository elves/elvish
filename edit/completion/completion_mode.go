package completion

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
)

// Completion mode.

// Interface.

type completion struct {
	binding      eddefs.BindingMap
	matcher      hashmap.Map
	argCompleter hashmap.Map
	completionState
}

type completionState struct {
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

func Init(ed eddefs.Editor, ns eval.Ns) {
	c := &completion{
		binding:      eddefs.EmptyBindingMap,
		matcher:      vals.MakeMapFromKV("", matchPrefix),
		argCompleter: makeArgCompleter(),
	}

	ns.AddNs("completion",
		eval.Ns{
			"binding":       vars.FromPtr(&c.binding),
			"matcher":       vars.FromPtr(&c.matcher),
			"arg-completer": vars.FromPtr(&c.argCompleter),
		}.AddBuiltinFns("edit:completion:", map[string]interface{}{
			"start":          func() { c.start(ed, false) },
			"smart-start":    func() { c.start(ed, true) },
			"up":             func() { c.prev(false) },
			"up-cycle":       func() { c.prev(true) },
			"down":           func() { c.next(false) },
			"down-cycle":     func() { c.next(true) },
			"left":           c.left,
			"right":          c.right,
			"accept":         func() { c.accept(ed) },
			"trigger-filter": c.triggerFilter,
			"default":        func() { c.complDefault(ed) },
		}))

	// Exposing arg completers.
	for _, v := range argCompletersData {
		ns[v.name+eval.FnSuffix] = vars.NewRo(
			&builtinArgCompleter{v.name, v.impl, c.argCompleter})
	}

	// Matchers.
	ns.AddFn("match-prefix", matchPrefix)
	ns.AddFn("match-substr", matchSubstr)
	ns.AddFn("match-subseq", matchSubseq)

	// Other functions.
	ns.AddBuiltinFns("edit:", map[string]interface{}{
		"complete-getopt":   complGetopt,
		"complex-candidate": makeComplexCandidate,
	})
}

func makeArgCompleter() hashmap.Map {
	m := vals.EmptyMap
	for k, v := range argCompletersData {
		m = m.Assoc(k, &builtinArgCompleter{v.name, v.impl, m})
	}
	return m
}

func (c *completion) Teardown() {
	c.completionState = completionState{}
}

func (c *completion) Binding(k ui.Key) eval.Callable {
	return c.binding.GetOrDefault(k)
}

func (c *completion) Replacement() (int, int, string) {
	return c.begin, c.end, c.selectedCandidate().code
}

func (*completion) RedrawModeLine() {}

func (c *completion) needScrollbar() bool {
	return c.firstShown > 0 || c.lastShownInFull < len(c.filtered)-1
}

func (c *completion) ModeLine() ui.Renderer {
	ml := ui.NewModeLineRenderer(
		fmt.Sprintf(" COMPLETING %s ", c.completer), c.filter)
	if !c.needScrollbar() {
		return ml
	}
	return ui.NewModeLineWithScrollBarRenderer(ml,
		len(c.filtered), c.firstShown, c.lastShownInFull+1)
}

func (c *completion) CursorOnModeLine() bool {
	return c.filtering
}

func (c *completion) left() {
	if x := c.selected - c.height; x >= 0 {
		c.selected = x
	}
}

func (c *completion) right() {
	if x := c.selected + c.height; x < len(c.filtered) {
		c.selected = x
	}
}

// acceptCompletion accepts currently selected completion candidate.
func (c *completion) accept(ed eddefs.Editor) {
	if 0 <= c.selected && c.selected < len(c.filtered) {
		ed.SetBuffer(c.apply(ed.Buffer()))
	}
	ed.SetModeInsert()
}

func (c *completion) complDefault(ed eddefs.Editor) {
	k := ed.LastKey()
	if c.filtering && likeChar(k) {
		c.changeFilter(c.filter + string(k.Rune))
	} else if c.filtering && k == (ui.Key{ui.Backspace, 0}) {
		_, size := utf8.DecodeLastRuneInString(c.filter)
		if size > 0 {
			c.changeFilter(c.filter[:len(c.filter)-size])
		}
	} else {
		c.accept(ed)
		ed.SetAction(eddefs.ReprocessKey)
	}
}

func (c *completion) triggerFilter() {
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

func (c *completion) start(ed eddefs.Editor, acceptSingleton bool) {
	_, dot := ed.Buffer()
	chunk := ed.ParsedBuffer()
	node := parseutil.FindLeafNode(chunk, dot)
	if node == nil {
		return
	}

	completer, complSpec, err := complete(
		node, &complEnv{ed.Evaler(), c.matcher, c.argCompleter})

	if err != nil {
		ed.AddTip("%v", err)
		// We don't show the full stack trace. To make debugging still possible,
		// we log it.
		if pprinter, ok := err.(util.Pprinter); ok {
			logger.Println("matcher error:")
			logger.Println(pprinter.Pprint(""))
		}
	} else if completer == "" {
		ed.AddTip("unsupported completion :(")
		logger.Println("path to current leaf, leaf first")
		for n := node; n != nil; n = n.Parent() {
			logger.Printf("%T (%d-%d)", n, n.Begin(), n.End())
		}
	} else if len(complSpec.candidates) == 0 {
		ed.AddTip("no candidate for %s", completer)
	} else {
		if acceptSingleton && len(complSpec.candidates) == 1 {
			// Just accept this single candidate.
			repl := complSpec.candidates[0].code
			buffer, _ := ed.Buffer()
			ed.SetBuffer(
				buffer[:complSpec.begin]+repl+buffer[complSpec.end:],
				complSpec.begin+len(repl))
			return
		}
		c.completionState = completionState{
			completer: completer,
			complSpec: *complSpec,
			filtered:  complSpec.candidates,
		}
		ed.SetMode(c)
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
	bb := ui.NewBufferBuilder(width)
	cands := c.filtered
	if len(cands) == 0 {
		bb.WriteString(util.TrimWcwidth("(no result)", width), "")
		return bb.Buffer()
	}
	if maxHeight <= 1 || width <= 2 {
		bb.WriteString(util.TrimWcwidth("(terminal too small)", width), "")
		return bb.Buffer()
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

		col := ui.NewBufferBuilder(totalColWidth)
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

		bb.ExtendRight(col.Buffer(), 0)
		remainedWidth -= totalColWidth
		if remainedWidth <= completionColMarginTotal {
			break
		}
	}
	// When the listing is incomplete, always use up the entire width.
	if remainedWidth > 0 && c.needScrollbar() {
		col := ui.NewBufferBuilder(remainedWidth)
		for i := 0; i < height; i++ {
			if i > 0 {
				col.Newline()
			}
			col.WriteSpaces(remainedWidth, styleForCompletion.String())
		}
		bb.ExtendRight(col.Buffer(), 0)
		remainedWidth = 0
	}
	return bb.Buffer()
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
