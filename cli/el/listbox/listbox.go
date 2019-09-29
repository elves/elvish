// Package listbox implements a widget for displaying and navigating a list of
// items.
package listbox

import (
	"strings"
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

// Widget supports displaying a list of items, including scrolling and selecting
// functions. It implements the clitypes.Widget interface. An empty Widget is
// directly usable.
type Widget struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	// Publically accessible state.
	State State

	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
	// A placeholder to show when there are no items.
	Placeholder styled.Text
	// A function to call when the selected item has changed.
	OnSelect func(it Items, i int)
	// A function called on the accept event.
	OnAccept func(it Items, i int)
	// Whether the listbox should be rendered in a horizontal layout. Note that
	// in the horizontal layout, items must have only one line.
	Horizontal bool
	// The minimal amount of space to reserve for left and right sides of each
	// entry.
	Padding int
}

var _ = el.Widget(&Widget{})

func (w *Widget) init() {
	if w.OverlayHandler == nil {
		w.OverlayHandler = el.DummyHandler{}
	}
	if w.OnSelect == nil {
		w.OnSelect = func(Items, int) {}
	}
	if w.OnAccept == nil {
		w.OnAccept = func(Items, int) {}
	}
}

var styleForSelected = "inverse"

func (w *Widget) Render(width, height int) *ui.Buffer {
	w.init()
	if w.Horizontal {
		return w.renderHorizontal(width, height)
	}
	return w.renderVertical(width, height)
}

const colGap = 2

func (w *Widget) renderHorizontal(width, height int) *ui.Buffer {
	var state State
	var allFit bool
	w.MutateListboxState(func(s *State) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, allFit = getHorizontalWindow(*s, width, height)
		}
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return layout.Label{Content: w.Placeholder}.Render(width, height)
	}

	items, selected, first := state.Items, state.Selected, state.First
	n := items.Len()

	buf := ui.NewBuffer(0)
	remainedWidth := width
	hasCropped := false
	last := first
	if !allFit {
		// Reserve one line for the scrollbar.
		height--
	}
	for i := first; i < n; i += height {
		// Render the column starting from i.
		col := make([]styled.Text, 0, height)
		for j := i; j < i+height && j < n; j++ {
			last = j
			item := items.Show(j)
			if j == selected {
				item = styled.Transform(item, styleForSelected)
			}
			col = append(col, item)
		}

		colWidth := maxWidth(items, i, i+height)
		if colWidth > remainedWidth {
			colWidth = remainedWidth
			hasCropped = true
		}

		colBuf := layout.CroppedLines{Lines: col}.Render(colWidth, height)
		buf.ExtendRight(colBuf)

		remainedWidth -= colWidth
		if remainedWidth > colGap {
			remainedWidth -= colGap
			buf.Width += colGap
		} else {
			buf.Width = width
			break
		}
	}
	if first != 0 || last != n-1 || hasCropped {
		scrollbar := layout.HScrollbar{Total: n, Low: first, High: last + 1}
		buf.Extend(scrollbar.Render(width, 1), false)
	}
	return buf
}

func (w *Widget) renderVertical(width, height int) *ui.Buffer {
	var state State
	var firstCrop int
	w.MutateListboxState(func(s *State) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, firstCrop = getVertialWindow(*s, height)
		}
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return layout.Label{Content: w.Placeholder}.Render(width, height)
	}

	items, selected, first := state.Items, state.Selected, state.First
	n := items.Len()
	allLines := []styled.Text{}
	hasCropped := firstCrop > 0

	var i, selectFrom, selectTo int
	for i = first; i < n && len(allLines) < height; i++ {
		item := items.Show(i)
		lines := item.SplitByRune('\n')
		if i == first {
			lines = lines[firstCrop:]
		}
		if i == selected {
			selectFrom, selectTo = len(allLines), len(allLines)+len(lines)
			for j := range lines {
				lines[j] = styled.Transform(
					lines[j].ConcatText(styled.Plain(strings.Repeat(" ", width))),
					styleForSelected)
			}
		}
		// TODO: Optionally, add underlines to the last line as a visual
		// separator between adjacent entries.

		if len(allLines)+len(lines) > height {
			lines = lines[:len(allLines)+len(lines)-height]
			hasCropped = true
		}
		allLines = append(allLines, lines...)
	}

	var rd el.Renderer = croppedLines{
		lines: allLines, padding: w.Padding,
		selectFrom: selectFrom, selectTo: selectTo}
	if first > 0 || i < n || hasCropped {
		rd = layout.VScrollbarContainer{
			Content:   rd,
			Scrollbar: layout.VScrollbar{Total: n, Low: first, High: i},
		}
	}
	return rd.Render(width, height)
}

type croppedLines struct {
	lines      []styled.Text
	padding    int
	selectFrom int
	selectTo   int
}

func (c croppedLines) Render(width, height int) *ui.Buffer {
	lines := c.lines
	bb := ui.NewBufferBuilder(width)
	padding := strings.Repeat(" ", c.padding)
	selectedPadding := styled.MakeText(padding, styleForSelected)
	for i, line := range lines {
		if i > 0 {
			bb.Newline()
		}
		if c.selectFrom <= i && i < c.selectTo {
			bb.WriteStyled(selectedPadding)
			bb.WriteStyled(
				styled.Transform(line, styleForSelected).
					ConcatText(styled.Plain(
						strings.Repeat(" ", width-2*c.padding))).
					TrimWcwidth(width - 2*c.padding))
			bb.WriteStyled(selectedPadding)
		} else {
			bb.WritePlain(padding)
			bb.WriteStyled(line.TrimWcwidth(width - 2*c.padding))
		}
	}
	return bb.Buffer()
}

func (w *Widget) Handle(event term.Event) bool {
	w.init()

	if w.OverlayHandler.Handle(event) {
		return true
	}

	switch event {
	case term.K(ui.Up):
		w.Select(Prev)
		return true
	case term.K(ui.Down):
		w.Select(Next)
		return true
	case term.K(ui.Enter):
		w.Accept()
		return true
	}
	return false
}

// Select changes the selected item by calling the given callback with the index
// of the current selected item and the total number of items and using the
// return value as the index of the new selected item.
func (w *Widget) Select(f func(selected, n int) int) {
	w.init()
	var it Items
	var i int
	w.MutateListboxState(func(s *State) {
		s.Selected = f(s.Selected, s.Items.Len())
		it, i = s.Items, s.Selected
	})
	w.OnSelect(it, i)
}

// Prev moves the selection to the previous item, or does nothing if the
// first item is currently selected. It is a suitable as an argument to
// ((*Widget).Select.
func Prev(selected, n int) int {
	switch {
	case selected >= n:
		return n - 1
	case selected <= 0:
		return 0
	default:
		return selected - 1
	}
}

// PrevWrap moves the selection to the previous item, or to the last item if
// the first item is currently selected. It is a suitable as an argument to
// ((*Widget).Select.
func PrevWrap(selected, n int) int {
	switch {
	case selected >= n:
		return n - 1
	case selected <= 0:
		return n - 1
	default:
		return selected - 1
	}
}

// Next moves the selection to the previous item, or does nothing if the
// last item is currently selected. It is a suitable as an argument to
// ((*Widget).Select.
func Next(selected, n int) int {
	switch {
	case selected >= n-1:
		return n - 1
	case selected < 0:
		return 0
	default:
		return selected + 1
	}
}

// NextWrap moves the selection to the previous item, or to the first item
// if the last item is currently selected. It is a suitable as an argument to
// ((*Widget).Select.
func NextWrap(selected, n int) int {
	switch {
	case selected >= n-1:
		return 0
	case selected < 0:
		return 0
	default:
		return selected + 1
	}
}

// Accept accepts the currently selected item.
func (w *Widget) Accept() {
	state := w.CopyListboxState()
	w.OnAccept(state.Items, state.Selected)
}

// MutateListboxState calls the given function while locking StateMutex.
func (w *Widget) MutateListboxState(f func(*State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}

// CopyListboxState returns a copy of the state.
func (w *Widget) CopyListboxState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}
