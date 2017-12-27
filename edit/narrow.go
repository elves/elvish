package edit

import (
	"container/list"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/hashmap"
)

var _ = registerBuiltins(modeNarrow, map[string]func(*Editor){
	"up":         func(ed *Editor) { getNarrow(ed).up(false) },
	"up-cycle":   func(ed *Editor) { getNarrow(ed).up(true) },
	"page-up":    func(ed *Editor) { getNarrow(ed).pageUp() },
	"down":       func(ed *Editor) { getNarrow(ed).down(false) },
	"down-cycle": func(ed *Editor) { getNarrow(ed).down(true) },
	"page-down":  func(ed *Editor) { getNarrow(ed).pageDown() },
	"backspace":  func(ed *Editor) { getNarrow(ed).backspace() },
	"accept":     func(ed *Editor) { getNarrow(ed).accept(ed) },
	"accept-close": func(ed *Editor) {
		getNarrow(ed).accept(ed)
		insertStart(ed)
	},
	"toggle-ignore-duplication": func(ed *Editor) {
		l := getNarrow(ed)
		l.opts.IgnoreDuplication = !l.opts.IgnoreDuplication
		l.refresh()
	},
	"toggle-ignore-case": func(ed *Editor) {
		l := getNarrow(ed)
		l.opts.IgnoreCase = !l.opts.IgnoreCase
		l.refresh()
	},
	"default": func(ed *Editor) { getNarrow(ed).defaultBinding(ed) },
})

func init() {
	registerBindings(modeNarrow, modeNarrow, nil)
}

// narrow implements a listing mode that supports the notion of selecting an
// entry and filtering entries.
type narrow struct {
	name        string
	selected    int
	filter      string
	pagesize    int
	headerWidth int

	placehold string
	source    func() []narrowItem
	action    func(*Editor, narrowItem)
	match     func(string, string) bool
	filtered  []narrowItem
	opts      narrowOptions
}

func (l *narrow) Binding(m map[string]eval.Variable, k ui.Key) eval.CallableValue {
	if l.opts.bindingMap != nil {
		if f, ok := l.opts.bindingMap[k]; ok {
			return f
		}
	}

	return getBinding(m[modeNarrow], k)
}

func (l *narrow) ModeLine() ui.Renderer {
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
	return modeLineRenderer{ml, l.filter}
}

func (l *narrow) CursorOnModeLine() bool {
	return true
}

func (l *narrow) List(maxHeight int) ui.Renderer {
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

func (l *narrow) refresh() {
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

func (l *narrow) changeFilter(newfilter string) {
	l.filter = newfilter
	l.refresh()
}

func (l *narrow) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *narrow) up(cycle bool) {
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

func (l *narrow) pageUp() {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected -= l.pagesize
	if l.selected < 0 {
		l.selected = 0
	}
}

func (l *narrow) down(cycle bool) {
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

func (l *narrow) pageDown() {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected += l.pagesize
	if l.selected >= n {
		l.selected = n - 1
	}
}

func (l *narrow) accept(ed *Editor) {
	if l.selected >= 0 {
		l.action(ed, l.filtered[l.selected])
	}
}

func (l *narrow) handleFilterKey(ed *Editor) bool {
	k := ed.lastKey
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		if len(l.filtered) == 1 && l.opts.AutoCommit {
			l.accept(ed)
			insertStart(ed)
		}
		return true
	}
	return false
}

func (l *narrow) defaultBinding(ed *Editor) {
	if !l.handleFilterKey(ed) {
		insertStart(ed)
		ed.setAction(reprocessKey)
	}
}

var errNotNarrow = errors.New("not in a narrow mode")

func getNarrow(ed *Editor) *narrow {
	if l, ok := ed.mode.(*narrow); ok {
		return l
	} else {
		throw(errNotNarrow)
		panic("unreachable")
	}
}

type narrowItem interface {
	eval.Value
	Display() ui.Styled
	Content() string
	FilterText() string
}

type narrowOptions struct {
	AutoCommit        bool
	Bindings          eval.Map
	IgnoreDuplication bool
	IgnoreCase        bool
	KeepBottom        bool
	MaxLines          int
	Modeline          string

	bindingMap map[ui.Key]eval.CallableValue
}

type narrowItemString struct {
	eval.String
}

func (s *narrowItemString) Content() string {
	return string(s.String)
}

func (s *narrowItemString) Display() ui.Styled {
	return ui.Unstyled(string(s.String))
}

func (s *narrowItemString) FilterText() string {
	return s.Content()
}

type narrowItemComplex struct {
	eval.Map
}

func (c *narrowItemComplex) Content() string {
	key := eval.String("content")
	if !c.Map.HasKey(key) {
		return ""
	}
	if s, ok := c.Map.IndexOne(key).(eval.String); !ok {
		return ""
	} else {
		return s.String()
	}
}

// TODO: add style
func (c *narrowItemComplex) Display() ui.Styled {
	key := eval.String("display")
	if !c.Map.HasKey(key) {
		return ui.Unstyled("")
	}
	if s, ok := c.Map.IndexOne(key).(eval.String); !ok {
		return ui.Unstyled("")
	} else {
		return ui.Unstyled(s.String())
	}
}

func (c *narrowItemComplex) FilterText() string {
	key := eval.String("filter-text")
	if c.Map.HasKey(key) {
		return c.Map.IndexOne(key).(eval.String).String()
	}
	return c.Content()
}

func NarrowRead(ec *eval.Frame, args []eval.Value, opts map[string]eval.Value) {
	var source, action eval.CallableValue
	l := &narrow{
		opts: narrowOptions{
			Bindings: eval.NewMap(hashmap.Empty),
		},
	}

	eval.ScanArgs(args, &source, &action)
	eval.ScanOptsToStruct(opts, &l.opts)

	l.opts.Bindings.IterateKey(func(k eval.Value) bool {
		key := ui.ToKey(k)
		f := l.opts.Bindings.IndexOne(k)
		maybeThrow(eval.ShouldBeFn(f))
		if l.opts.bindingMap == nil {
			l.opts.bindingMap = make(map[ui.Key]eval.CallableValue)
		}
		l.opts.bindingMap[key] = f.(eval.CallableValue)
		return true
	})

	l.source = narrowGetSource(ec, source)
	l.action = func(ed *Editor, item narrowItem) {
		ed.CallFn(action, item)
	}
	// TODO: user customize varible
	l.match = strings.Contains

	l.changeFilter("")
	ed := ec.Editor.(*Editor)
	ed.mode = l
}

func narrowGetSource(ec *eval.Frame, source eval.CallableValue) func() []narrowItem {
	return func() []narrowItem {
		ed := ec.Editor.(*Editor)
		vs, err := ec.PCaptureOutput(source, eval.NoArgs, eval.NoOpts)
		if err != nil {
			ed.Notify(err.Error())
			return nil
		}
		var lis []narrowItem
		for _, v := range vs {
			switch raw := v.(type) {
			case eval.String:
				lis = append(lis, &narrowItemString{raw})
			case eval.Map:
				lis = append(lis, &narrowItemComplex{raw})
			}
		}
		return lis
	}
}

func CommandHistory(ec *eval.Frame, args []eval.Value, opts map[string]eval.Value) {
	var (
		rest              []int
		limit, start, end int
	)

	eval.ScanArgsVariadic(args, &rest)
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	ed := ec.Editor.(*Editor)
	cmds, err := ed.historyFuser.AllCmds()
	if err != nil {
		return
	}

	if len(rest) > 0 {
		limit = rest[0]
	}

	total := len(cmds)
	switch {
	case limit > 0:
		start = 0
		end = limit
		if limit > total {
			end = total
		}
	case limit < 0:
		start = limit + total
		if start < 0 {
			start = 0
		}
		end = total
	default:
		start = 0
		end = total
	}

	for i := start; i < end; i++ {
		out <- eval.ConvertToMap(map[eval.Value]eval.Value{
			eval.String("id"):  eval.String(strconv.Itoa(i)),
			eval.String("cmd"): eval.String(cmds[i]),
		})
	}
}

func InsertAtDot(ec *eval.Frame, args []eval.Value, opts map[string]eval.Value) {
	var text eval.String

	eval.ScanArgs(args, &text)
	eval.TakeNoOpt(opts)

	ed := ec.Editor.(*Editor)
	ed.insertAtDot(text.String())
}

func ReplaceInput(ec *eval.Frame, args []eval.Value, opts map[string]eval.Value) {
	var text eval.String

	eval.ScanArgs(args, &text)
	eval.TakeNoOpt(opts)

	ed := ec.Editor.(*Editor)
	ed.buffer = text.String()
}

func Wordify(ec *eval.Frame, args []eval.Value, opts map[string]eval.Value) {
	var text eval.String

	eval.ScanArgs(args, &text)
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	for _, s := range wordify(text.String()) {
		out <- eval.String(s)
	}
}
