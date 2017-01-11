package edit

import (
	"container/list"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/util"
)

// listing implements a listing mode that supports the notion of selecting an
// entry and filtering entries.
type listing struct {
	typ      ModeType
	provider listingProvider
	selected int
	filter   string
	pagesize int
}

type listingProvider interface {
	Len() int
	Show(i, w int) styled
	Filter(filter string) int
	Accept(i int, ed *Editor)
	ModeTitle(int) string
}

type Placeholderer interface {
	Placeholder() string
}

func newListing(t ModeType, p listingProvider) listing {
	l := listing{t, p, 0, "", 0}
	l.changeFilter("")
	return l
}

func (l *listing) Mode() ModeType {
	return l.typ
}

func (l *listing) ModeLine(width int) *buffer {
	title := l.provider.ModeTitle(l.selected)
	// TODO keep it one line.
	b := newBuffer(width)
	b.writes(util.TrimWcwidth(title, width), styleForMode.String())
	b.writes(" ", "")
	b.writes(l.filter, styleForFilter.String())
	b.dot = b.cursor()
	return b
}

func (l *listing) List(width, maxHeight int) *buffer {
	n := l.provider.Len()
	b := newBuffer(width)
	if n == 0 {
		var ph string
		if pher, ok := l.provider.(Placeholderer); ok {
			ph = pher.Placeholder()
		} else {
			ph = "(no result)"
		}
		b.writes(util.TrimWcwidth(ph, width), "")
		return b
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
	var lines list.List
	getEntry := func(i int) []styled {
		s := l.provider.Show(i, width)
		lines := strings.Split(s.text, "\n")
		st := s.styles
		if i == l.selected {
			st = append(st, styleForSelected.String())
		}
		styleds := make([]styled, len(lines))
		for i, line := range lines {
			styleds[i] = styled{line, st}
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
				lines.PushFront(entry[i])
				height++
			}
			if i >= 0 {
				lastShownIncomplete = true
			}
		} else {
			entry := getEntry(high)
			// Append at most the first (height - maxHeight) lines.
			for i = 0; i < len(entry) && height < maxHeight; i++ {
				lines.PushBack(entry[i])
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

	var scrollbar *buffer
	if low > 0 || high < n || lastShownIncomplete {
		scrollbar = renderScrollbar(n, low, high, height)
		width--
	}

	for p := lines.Front(); p != nil; p = p.Next() {
		s := p.Value.(styled)
		if p != lines.Front() {
			b.newline()
		}
		b.writes(s.text, s.styles.String())
	}

	if scrollbar != nil {
		b.extendHorizontal(scrollbar, width)
	}
	return b
}

func writeHorizontalScrollbar(b *buffer, n, low, high, width int) {
	slow, shigh := findScrollInterval(n, low, high, width)
	for i := 0; i < width; i++ {
		if slow <= i && i < shigh {
			b.write(' ', styleForScrollBarThumb.String())
		} else {
			b.write('━', styleForScrollBarArea.String())
		}
	}
}

func renderScrollbar(n, low, high, height int) *buffer {
	slow, shigh := findScrollInterval(n, low, high, height)
	// Logger.Printf("low = %d, high = %d, n = %d, slow = %d, shigh = %d", low, high, n, slow, shigh)
	b := newBuffer(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			b.newline()
		}
		if slow <= i && i < shigh {
			b.write(' ', styleForScrollBarThumb.String())
		} else {
			b.write('│', styleForScrollBarArea.String())
		}
	}
	return b
}

func findScrollInterval(n, low, high, height int) (int, int) {
	f := func(i int) int {
		return int(float64(i)/float64(n)*float64(height) + 0.5)
	}
	scrollLow, scrollHigh := f(low), f(high)
	if scrollLow == scrollHigh {
		if scrollHigh == high {
			scrollLow--
		} else {
			scrollHigh++
		}
	}
	return scrollLow, scrollHigh
}

func (l *listing) changeFilter(newfilter string) {
	l.filter = newfilter
	l.selected = l.provider.Filter(newfilter)
}

func (l *listing) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *listing) up(cycle bool) {
	n := l.provider.Len()
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

func (l *listing) pageUp() {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected -= l.pagesize
	if l.selected < 0 {
		l.selected = 0
	}
}

func (l *listing) down(cycle bool) {
	n := l.provider.Len()
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

func (l *listing) pageDown() {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected += l.pagesize
	if l.selected >= n {
		l.selected = n - 1
	}
}

func (l *listing) accept(ed *Editor) {
	if l.selected >= 0 {
		l.provider.Accept(l.selected, ed)
	}
}

func (l *listing) handleFilterKey(k Key) bool {
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		return true
	}
	return false
}

func (l *listing) defaultBinding(ed *Editor) {
	if !l.handleFilterKey(ed.lastKey) {
		startInsert(ed)
		ed.nextAction = action{typ: reprocessKey}
	}
}

func addListingBuiltins(prefix string, l func(*Editor) *listing) {
	add := func(name string, f func(*Editor)) {
		builtins = append(builtins, &BuiltinFn{prefix + name, f})
	}
	add("up", func(ed *Editor) { l(ed).up(false) })
	add("up-cycle", func(ed *Editor) { l(ed).up(true) })
	add("page-up", func(ed *Editor) { l(ed).pageUp() })
	add("down", func(ed *Editor) { l(ed).down(false) })
	add("down-cycle", func(ed *Editor) { l(ed).down(true) })
	add("page-down", func(ed *Editor) { l(ed).pageDown() })
	add("backspace", func(ed *Editor) { l(ed).backspace() })
	add("accept", func(ed *Editor) { l(ed).accept(ed) })
	add("default", func(ed *Editor) { l(ed).defaultBinding(ed) })
}

func addListingDefaultBindings(prefix string, m ModeType) {
	add := func(k Key, name string) {
		if _, ok := defaultBindings[m][k]; !ok {
			defaultBindings[m][k] = prefix + name
		}
	}
	add(Key{Up, 0}, "up")
	add(Key{PageUp, 0}, "page-up")
	add(Key{Down, 0}, "down")
	add(Key{PageDown, 0}, "page-down")
	add(Key{Tab, 0}, "down-cycle")
	add(Key{Backspace, 0}, "backspace")
	add(Key{Enter, 0}, "accept")
	add(Default, "default")
	defaultBindings[m][Key{'[', Ctrl}] = "start-insert"
}
