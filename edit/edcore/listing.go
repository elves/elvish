package edcore

import (
	"container/list"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/util"
)

// listingMode implements a mode that supports listing, selecting and filtering
// entries.
type listingMode struct {
	commonBinding eddefs.BindingMap
	listingState
}

type listingState struct {
	binding     eddefs.BindingMap
	provider    eddefs.ListingProvider
	selected    int
	filter      string
	pagesize    int
	headerWidth int
}

func init() { atEditorInit(initListing) }

func initListing(ed *editor, ns eval.Ns) {
	l := &listingMode{commonBinding: emptyBindingMap}
	ed.listing = l

	subns := eval.Ns{
		"binding": vars.FromPtr(&l.commonBinding),
	}
	subns.AddBuiltinFns("edit:listing:", map[string]interface{}{
		"up":         func() { l.up(false) },
		"up-cycle":   func() { l.up(true) },
		"page-up":    func() { l.pageUp() },
		"down":       func() { l.down(false) },
		"down-cycle": func() { l.down(true) },
		"page-down":  func() { l.pageDown() },
		"backspace":  func() { l.backspace() },
		"accept":     func() { l.accept(ed) },
		"accept-close": func() {
			l.accept(ed)
			ed.SetModeInsert()
		},
		"default": func() { l.defaultBinding(ed) },
	})
	ns.AddNs("listing", subns)
}

type placeholderer interface {
	Placeholder() string
}

func (l *listingMode) Teardown() {
	l.listingState = listingState{}
	if p, ok := l.provider.(teardowner); ok {
		p.Teardown()
	}
}

type teardowner interface {
	Teardown()
}

func (l *listingMode) Binding(k ui.Key) eval.Callable {
	specificBindings := l.binding
	listingBindings := l.commonBinding
	// mode-specific binding -> listing binding ->
	// mode-specific default -> listing default
	switch {
	case specificBindings.HasKey(k):
		return specificBindings.GetKey(k)
	case listingBindings.HasKey(k):
		return listingBindings.GetKey(k)
	case specificBindings.HasKey(ui.Default):
		return specificBindings.GetKey(ui.Default)
	case listingBindings.HasKey(ui.Default):
		return listingBindings.GetKey(ui.Default)
	default:
		return nil
	}
}

func newListing(b eddefs.BindingMap, p eddefs.ListingProvider) *listingState {
	l := &listingState{}
	l.setup(b, p)
	return l
}

func (l *listingState) setup(b eddefs.BindingMap, p eddefs.ListingProvider) {
	*l = listingState{b, p, 0, "", 0, 0}
	l.refresh()
	for i := 0; i < p.Len(); i++ {
		header, _ := p.Show(i)
		width := util.Wcswidth(header)
		if l.headerWidth < width {
			l.headerWidth = width
		}
	}
}

func (l *listingState) ModeLine() ui.Renderer {
	return ui.NewModeLineRenderer(l.provider.ModeTitle(l.selected), l.filter)
}

func (l *listingState) CursorOnModeLine() bool {
	if c, ok := l.provider.(cursorOnModeLiner); ok {
		return c.CursorOnModeLine()
	}
	return false
}

func (l *listingState) List(maxHeight int) ui.Renderer {
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

func renderScrollbar(n, low, high, height int) *ui.Buffer {
	slow, shigh := findScrollInterval(n, low, high, height)
	// Logger.Printf("low = %d, high = %d, n = %d, slow = %d, shigh = %d", low, high, n, slow, shigh)
	bb := ui.NewBufferBuilder(1)
	for i := 0; i < height; i++ {
		if i > 0 {
			bb.Newline()
		}
		if slow <= i && i < shigh {
			bb.Write(' ', styleForScrollBarThumb.String())
		} else {
			bb.Write('│', styleForScrollBarArea.String())
		}
	}
	return bb.Buffer()
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

func (l *listingState) changeFilter(newfilter string) {
	l.filter = newfilter
	l.refresh()
}

func (l *listingState) refresh() {
	l.selected = l.provider.Filter(l.filter)
}

func (l *listingState) backspace() bool {
	_, size := utf8.DecodeLastRuneInString(l.filter)
	if size > 0 {
		l.changeFilter(l.filter[:len(l.filter)-size])
		return true
	}
	return false
}

func (l *listingState) up(cycle bool) {
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

func (l *listingState) pageUp() {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected -= l.pagesize
	if l.selected < 0 {
		l.selected = 0
	}
}

func (l *listingState) down(cycle bool) {
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

func (l *listingState) pageDown() {
	n := l.provider.Len()
	if n == 0 {
		return
	}
	l.selected += l.pagesize
	if l.selected >= n {
		l.selected = n - 1
	}
}

func (l *listingState) accept(ed *editor) {
	if l.selected >= 0 {
		l.provider.Accept(l.selected, ed)
	}
}

func (l *listingState) defaultBinding(ed *editor) {
	k := ed.LastKey()
	if likeChar(k) {
		// Append to filter
		l.changeFilter(l.filter + string(k.Rune))
		if aa, ok := l.provider.(autoAccepter); ok {
			if aa.AutoAccept() {
				l.accept(ed)
			}
		}
	} else {
		ed.SetModeInsert()
		ed.SetAction(reprocessKey)
	}
}

type autoAccepter interface {
	AutoAccept() bool
}
