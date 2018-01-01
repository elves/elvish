package edit

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var _ = registerBuiltins(modeListing, map[string]func(*Editor){
	"up":         func(ed *Editor) { getListing(ed).up(false) },
	"up-cycle":   func(ed *Editor) { getListing(ed).up(true) },
	"page-up":    func(ed *Editor) { getListing(ed).pageUp() },
	"down":       func(ed *Editor) { getListing(ed).down(false) },
	"down-cycle": func(ed *Editor) { getListing(ed).down(true) },
	"page-down":  func(ed *Editor) { getListing(ed).pageDown() },
	"backspace":  func(ed *Editor) { getListing(ed).backspace() },
	"accept":     func(ed *Editor) { getListing(ed).accept(ed) },
	"accept-close": func(ed *Editor) {
		getListing(ed).accept(ed)
		insertStart(ed)
	},
	"default": func(ed *Editor) { getListing(ed).defaultBinding(ed) },
})

func init() {
	registerBindings(modeListing, modeListing, map[ui.Key]string{
		{ui.Up, 0}:         "up",
		{ui.PageUp, 0}:     "page-up",
		{ui.Down, 0}:       "down",
		{ui.PageDown, 0}:   "page-down",
		{ui.Tab, 0}:        "down-cycle",
		{ui.Tab, ui.Shift}: "up-cycle",
		{ui.Backspace, 0}:  "backspace",
		{ui.Enter, 0}:      "accept-close",
		{ui.Enter, ui.Alt}: "accept",
		ui.Default:         "default",
		{'[', ui.Ctrl}:     "insert:start",
	})
}

// listing implements a listing mode that supports the notion of selecting an
// entry and filtering entries.
type listing struct {
	name        string
	provider    listingProvider
	selected    int
	filter      string
	pagesize    int
	headerWidth int
}

type listingProvider interface {
	Len() int
	Show(i int) (string, ui.Styled)
	Filter(filter string) int
	Accept(i int, ed *Editor)
	ModeTitle(int) string
}

type placeholderer interface {
	Placeholder() string
}

func newListing(t string, p listingProvider) listing {
	l := listing{t, p, 0, "", 0, 0}
	l.refresh()
	for i := 0; i < p.Len(); i++ {
		header, _ := p.Show(i)
		width := util.Wcswidth(header)
		if l.headerWidth < width {
			l.headerWidth = width
		}
	}
	return l
}

func (l *listing) Binding(m map[string]eval.Variable, k ui.Key) eval.Fn {
	if m[l.name] == nil {
		return getBinding(m[modeListing], k)
	}
	specificBindings := m[l.name].Get().(BindingTable)
	listingBindings := m[modeListing].Get().(BindingTable)
	// mode-specific binding -> listing binding ->
	// mode-specific default -> listing default
	switch {
	case specificBindings.HasKey(k):
		return specificBindings.get(k)
	case listingBindings.HasKey(k):
		return listingBindings.get(k)
	case specificBindings.HasKey(ui.Default):
		return specificBindings.get(ui.Default)
	case listingBindings.HasKey(ui.Default):
		return listingBindings.get(ui.Default)
	default:
		return nil
	}
}

func (l *listing) ModeLine() ui.Renderer {
	return modeLineRenderer{l.provider.ModeTitle(l.selected), l.filter}
}

func (l *listing) CursorOnModeLine() bool {
	if c, ok := l.provider.(CursorOnModeLiner); ok {
		return c.CursorOnModeLine()
	}
	return false
}

func (l *listing) List(maxHeight int) ui.Renderer {
	n := l.provider.Len()
	if n == 0 {
		var ph string
		if pher, ok := l.provider.(placeholderer); ok {
			ph = pher.Placeholder()
		} else {
			ph = "(no result)"
		}
		return placeholderRenderer(ph)
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
		header, content := l.provider.Show(i)
		lines := strings.Split(content.Text, "\n")
		styles := content.Styles
		if i == l.selected {
			styles = append(styles, styleForSelected...)
		}
		styleds := make([]ui.Styled, len(lines))
		for i, line := range lines {
			if l.headerWidth > 0 {
				if i == 0 {
					line = fmt.Sprintf("%*s %s", l.headerWidth, header, line)
				} else {
					line = fmt.Sprintf("%*s %s", l.headerWidth, "", line)
				}
			}
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

func writeHorizontalScrollbar(b *ui.Buffer, n, low, high, width int) {
	slow, shigh := findScrollInterval(n, low, high, width)
	for i := 0; i < width; i++ {
		if slow <= i && i < shigh {
			b.Write(' ', styleForScrollBarThumb.String())
		} else {
			b.Write('━', styleForScrollBarArea.String())
		}
	}
}

func renderScrollbar(n, low, high, height int) *ui.Buffer {
	slow, shigh := findScrollInterval(n, low, high, height)
	// Logger.Printf("low = %d, high = %d, n = %d, slow = %d, shigh = %d", low, high, n, slow, shigh)
	b := ui.NewBuffer(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			b.Newline()
		}
		if slow <= i && i < shigh {
			b.Write(' ', styleForScrollBarThumb.String())
		} else {
			b.Write('│', styleForScrollBarArea.String())
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
	l.refresh()
}

func (l *listing) refresh() {
	l.selected = l.provider.Filter(l.filter)
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

func (l *listing) handleFilterKey(k ui.Key) bool {
	if likeChar(k) {
		l.changeFilter(l.filter + string(k.Rune))
		return true
	}
	return false
}

func (l *listing) defaultBinding(ed *Editor) {
	if !l.handleFilterKey(ed.lastKey) {
		insertStart(ed)
		ed.setAction(reprocessKey)
	}
}

var errNotListing = errors.New("not in a listing mode")

func getListing(ed *Editor) *listing {
	if l, ok := ed.mode.(*listing); ok {
		return l
	} else {
		throw(errNotListing)
		panic("unreachable")
	}
}
