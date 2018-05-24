package edcore

import (
	"container/list"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse/parseutil"
	"github.com/xiaq/persistent/hashmap"
)

// narrow implements a listing mode that supports the notion of selecting an
// entry and filtering entries.
type narrow struct {
	binding eddefs.BindingMap
	narrowState
}

func init() { atEditorInit(initNarrow) }

func initNarrow(ed *editor, ns eval.Ns) {
	n := &narrow{binding: emptyBindingMap}
	subns := eval.Ns{
		"binding": vars.FromPtr(&n.binding),
	}
	subns.AddBuiltinFns("edit:narrow:", map[string]interface{}{
		"up":         func() { n.up(false) },
		"up-cycle":   func() { n.up(true) },
		"page-up":    func() { n.pageUp() },
		"down":       func() { n.down(false) },
		"down-cycle": func() { n.down(true) },
		"page-down":  func() { n.pageDown() },
		"backspace":  func() { n.backspace() },
		"accept":     func() { n.accept(ed) },
		"accept-close": func() {
			n.accept(ed)
			ed.SetModeInsert()
		},
		"toggle-ignore-duplication": func() {
			n.opts.IgnoreDuplication = !n.opts.IgnoreDuplication
			n.refresh()
		},
		"toggle-ignore-case": func() {
			n.opts.IgnoreCase = !n.opts.IgnoreCase
			n.refresh()
		},
		"default": func() { n.defaultBinding(ed) },
	})
	ns.AddNs("narrow", subns)
	ns.AddBuiltinFn("edit:", "-narrow-read", n.NarrowRead)
}

type narrowState struct {
	name        string
	selected    int
	filter      string
	pagesize    int
	headerWidth int

	placehold string
	source    func() []narrowItem
	action    func(*editor, narrowItem)
	match     func(string, string) bool
	filtered  []narrowItem
	opts      narrowOptions
}

func (l *narrow) Teardown() {
	l.narrowState = narrowState{}
}

func (l *narrow) Binding(k ui.Key) eval.Callable {
	if l.opts.bindingMap != nil {
		if f, ok := l.opts.bindingMap[k]; ok {
			return f
		}
	}
	return l.binding.GetOrDefault(k)
}

func (l *narrowState) ModeLine() ui.Renderer {
	ml := l.opts.Modeline
	var opt []string
	if l.opts.AutoCommit {
		opt = append(opt, "A")
	}
	if l.opts.IgnoreCase {
		opt = append(opt, "C")
	}
	if l.opts.IgnoreDuplication {
		opt = append(opt, "D")
	}
	if len(opt) != 0 {
		ml += "[" + strings.Join(opt, " ") + "]"
	}
	return ui.NewModeLineRenderer(ml, l.filter)
}

func (l *narrowState) CursorOnModeLine() bool {
	return true
}

func (l *narrowState) List(maxHeight int) ui.Renderer {
	if l.opts.MaxLines > 0 && l.opts.MaxLines < maxHeight {
		maxHeight = l.opts.MaxLines
	}

	if l.filtered == nil {
		l.refresh()
	}
	n := len(l.filtered)
	if n == 0 {
		return placeholderRenderer(l.placehold)
	}

	// Collect the entries to show. We start from the selected entry and extend
	// in both directions alternatingly. The entries are split into lines and
	// then collected in a list.
	low := l.selected
	if low == -1 {
		low = 0
	}
	high := low
	height := 0
	var listOfLines list.List
	getEntry := func(i int) []ui.Styled {
		display := l.filtered[i].Display()
		lines := strings.Split(display.Text, "\n")
		styles := display.Styles
		if i == l.selected {
			styles = append(styles, styleForSelected...)
		}
		styleds := make([]ui.Styled, len(lines))
		for i, line := range lines {
			styleds[i] = ui.Styled{line, styles}
		}
		return styleds
	}
	// We start by extending high, so that the first entry to include is
	// l.selected.
	extendLow := false
	lastShownIncomplete := false
	for height < maxHeight && !(low == 0 && high == n) {
		var i int
		if (extendLow && low > 0) || high == n {
			low--

			entry := getEntry(low)
			// Prepend at most the last (height - maxHeight) lines.
			for i = len(entry) - 1; i >= 0 && height < maxHeight; i-- {
				listOfLines.PushFront(entry[i])
				height++
			}
			if i >= 0 {
				lastShownIncomplete = true
			}
		} else {
			entry := getEntry(high)
			// Append at most the first (height - maxHeight) lines.
			for i = 0; i < len(entry) && height < maxHeight; i++ {
				listOfLines.PushBack(entry[i])
				height++
			}
			if i < len(entry) {
				lastShownIncomplete = true
			}

			high++
		}
		extendLow = !extendLow
	}

	l.pagesize = high - low

	// Convert the List to a slice.
	lines := make([]ui.Styled, 0, listOfLines.Len())
	for p := listOfLines.Front(); p != nil; p = p.Next() {
		lines = append(lines, p.Value.(ui.Styled))
	}

	ls := listingRenderer{lines}
	if low > 0 || high < n || lastShownIncomplete {
		// Need scrollbar
		return listingWithScrollBarRenderer{ls, n, low, high, height}
	}
	return ls
}

func (l *narrowState) refresh() {
	var candidates []narrowItem
	if l.source != nil {
		candidates = l.source()
	}
	l.filtered = make([]narrowItem, 0, len(candidates))

	filter := l.filter
	if l.opts.IgnoreCase {
		filter = strings.ToLower(filter)
	}

	set := make(map[string]struct{})

	for _, item := range candidates {
		text := item.FilterText()
		s := text
		if l.opts.IgnoreCase {
			s = strings.ToLower(s)
		}
		if !l.match(s, filter) {
			continue
		}
		if l.opts.IgnoreDuplication {
			if _, ok := set[text]; ok {
				continue
			}
			set[text] = struct{}{}
		}
		l.filtered = append(l.filtered, item)
	}

	if l.opts.KeepBottom {
		l.selected = len(l.filtered) - 1
	} else {
		l.selected = 0
	}
}

func (l *narrowState) changeFilter(newfilter string) {
	l.filter = newfilter
	l.refresh()
}

func (l *narrowState) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *narrowState) up(cycle bool) {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected--
	if l.selected == -1 {
		if cycle {
			l.selected += n
		} else {
			l.selected++
		}
	}
}

func (l *narrowState) pageUp() {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected -= l.pagesize
	if l.selected < 0 {
		l.selected = 0
	}
}

func (l *narrowState) down(cycle bool) {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected++
	if l.selected == n {
		if cycle {
			l.selected -= n
		} else {
			l.selected--
		}
	}
}

func (l *narrowState) pageDown() {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected += l.pagesize
	if l.selected >= n {
		l.selected = n - 1
	}
}

func (l *narrowState) accept(ed *editor) {
	if l.selected >= 0 {
		l.action(ed, l.filtered[l.selected])
	}
}

func (l *narrowState) handleFilterKey(ed *editor) bool {
	k := ed.lastKey
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		if len(l.filtered) == 1 && l.opts.AutoCommit {
			l.accept(ed)
			ed.SetModeInsert()
		}
		return true
	}
	return false
}

func (l *narrowState) defaultBinding(ed *editor) {
	if !l.handleFilterKey(ed) {
		ed.SetModeInsert()
		ed.SetAction(reprocessKey)
	}
}

type narrowItem interface {
	Display() ui.Styled
	Content() string
	FilterText() string
}

type narrowOptions struct {
	AutoCommit        bool
	Bindings          hashmap.Map
	IgnoreDuplication bool
	IgnoreCase        bool
	KeepBottom        bool
	MaxLines          int
	Modeline          string

	bindingMap map[ui.Key]eval.Callable
}

type narrowItemString struct {
	String string
}

func (s *narrowItemString) Content() string {
	return s.String
}

func (s *narrowItemString) Display() ui.Styled {
	return ui.Unstyled(s.String)
}

func (s *narrowItemString) FilterText() string {
	return s.Content()
}

type narrowItemComplex struct {
	hashmap.Map
}

func (c *narrowItemComplex) Content() string {
	if v, ok := c.Map.Index("content"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// TODO: add style
func (c *narrowItemComplex) Display() ui.Styled {
	if v, ok := c.Map.Index("display"); ok {
		if s, ok := v.(string); ok {
			return ui.Unstyled(s)
		}
	}
	return ui.Unstyled("")
}

func (c *narrowItemComplex) FilterText() string {
	if v, ok := c.Map.Index("filter-text"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return c.Content()
}

func (n *narrow) NarrowRead(fm *eval.Frame, opts eval.RawOptions, source, action eval.Callable) {
	l := &narrowState{
		opts: narrowOptions{
			Bindings: vals.EmptyMap,
		},
	}

	opts.Scan(&l.opts)

	for it := l.opts.Bindings.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		key := ui.ToKey(k)
		val, ok := v.(eval.Callable)
		if !ok {
			throwf("should be fn")
		}
		if l.opts.bindingMap == nil {
			l.opts.bindingMap = make(map[ui.Key]eval.Callable)
		}
		l.opts.bindingMap[key] = val
	}

	l.source = narrowGetSource(fm, source)
	l.action = func(ed *editor, item narrowItem) {
		ed.CallFn(action, item)
	}
	// TODO: user customize varible
	l.match = strings.Contains

	l.changeFilter("")

	n.narrowState = *l
	fm.Editor.(*editor).SetMode(n)
}

func narrowGetSource(ec *eval.Frame, source eval.Callable) func() []narrowItem {
	return func() []narrowItem {
		ed := ec.Editor.(*editor)
		vs, err := ec.CaptureOutput(source, eval.NoArgs, eval.NoOpts)
		if err != nil {
			ed.Notify(err.Error())
			return nil
		}
		var lis []narrowItem
		for _, v := range vs {
			switch raw := v.(type) {
			case string:
				lis = append(lis, &narrowItemString{raw})
			case hashmap.Map:
				lis = append(lis, &narrowItemComplex{raw})
			}
		}
		return lis
	}
}

func (ed *editor) replaceInput(text string) {
	ed.buffer = text
}

func wordifyBuiltin(fm *eval.Frame, text string) {
	out := fm.OutputChan()
	for _, s := range parseutil.Wordify(text) {
		out <- s
	}
}
