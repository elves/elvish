// Package listbox implements a widget for displaying and navigating a list of
// items.
package listbox

import (
	"strings"
	"sync"

	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/ui"
)

// Widget supports displaying a list of items, including scrolling and selecting
// functions. It implements the clitypes.widget interface. An empty widget is
// directly usable.
type Widget interface {
	el.Widget
	// CopyState returns a copy of the state.
	CopyState() State
	// Reset resets the state of the widget with the given items and index of
	// the selected item. It triggers the OnSelect callback if the index is
	// valid.
	Reset(it Items, selected int)
	// Select changes the selection by calling f with the current state, and
	// using the return value as the new selection index. It triggers the
	// OnSelect callback if the selected index has changed and is valid.
	Select(f func(State) int)
	// Accept accepts the currently selected item.
	Accept()
}

// Spec specifies the configuration and initial state for Widget.
type Spec struct {
	// A Handler that takes precedence over the default handling of events.
	OverlayHandler el.Handler
	// A placeholder to show when there are no items.
	Placeholder ui.Text
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
	// If true, the left padding of each item will be styled the same as the
	// first segment of the item, and the right spacing and padding will be
	// styled the same as the last segment of the item.
	ExtendStyle bool

	// State. When used in New, this field specifies the initial state.
	State State
}

type widget struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	// Configuration and state.
	Spec
}

// New creates a new listbox widget from the given specification.
func New(spec Spec) Widget {
	if spec.OverlayHandler == nil {
		spec.OverlayHandler = el.DummyHandler{}
	}
	if spec.OnAccept == nil {
		spec.OnAccept = func(Items, int) {}
	}
	if spec.OnSelect == nil {
		spec.OnSelect = func(Items, int) {}
	} else {
		s := spec.State
		if s.Items != nil && 0 <= s.Selected && s.Selected < s.Items.Len() {
			spec.OnSelect(s.Items, s.Selected)
		}
	}
	return &widget{Spec: spec}
}

var styleForSelected = ui.Inverse

func (w *widget) Render(width, height int) *term.Buffer {
	if w.Horizontal {
		return w.renderHorizontal(width, height)
	}
	return w.renderVertical(width, height)
}

const colGap = 2

func (w *widget) renderHorizontal(width, height int) *term.Buffer {
	var state State
	w.mutate(func(s *State) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, s.Height = getHorizontalWindow(*s, w.Padding, width, height)
			// Override height to the height required; we don't need the
			// original height later.
			height = s.Height
		}
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return layout.Label{Content: w.Placeholder}.Render(width, height)
	}

	items, selected, first := state.Items, state.Selected, state.First
	n := items.Len()

	buf := term.NewBuffer(0)
	remainedWidth := width
	hasCropped := false
	last := first
	for i := first; i < n; i += height {
		selectedRow := -1
		// Render the column starting from i.
		col := make([]ui.Text, 0, height)
		for j := i; j < i+height && j < n; j++ {
			last = j
			item := items.Show(j)
			if j == selected {
				selectedRow = j - i
			}
			col = append(col, item)
		}

		colWidth := maxWidth(items, w.Padding, i, i+height)
		if colWidth > remainedWidth {
			colWidth = remainedWidth
			hasCropped = true
		}

		colBuf := croppedLines{
			lines: col, padding: w.Padding,
			selectFrom: selectedRow, selectTo: selectedRow + 1,
			extendStyle: w.ExtendStyle}.Render(colWidth, height)
		buf.ExtendRight(colBuf)

		remainedWidth -= colWidth
		if remainedWidth <= colGap {
			break
		}
		remainedWidth -= colGap
		buf.Width += colGap
	}
	// We may not have used all the width required; force buffer width.
	buf.Width = width
	if first != 0 || last != n-1 || hasCropped {
		scrollbar := layout.HScrollbar{Total: n, Low: first, High: last + 1}
		buf.Extend(scrollbar.Render(width, 1), false)
	}
	return buf
}

func (w *widget) renderVertical(width, height int) *term.Buffer {
	var state State
	var firstCrop int
	w.mutate(func(s *State) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, firstCrop = getVerticalWindow(*s, height)
		}
		s.Height = height
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return layout.Label{Content: w.Placeholder}.Render(width, height)
	}

	items, selected, first := state.Items, state.Selected, state.First
	n := items.Len()
	allLines := []ui.Text{}
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
		selectFrom: selectFrom, selectTo: selectTo, extendStyle: w.ExtendStyle}
	if first > 0 || i < n || hasCropped {
		rd = layout.VScrollbarContainer{
			Content:   rd,
			Scrollbar: layout.VScrollbar{Total: n, Low: first, High: i},
		}
	}
	return rd.Render(width, height)
}

type croppedLines struct {
	lines       []ui.Text
	padding     int
	selectFrom  int
	selectTo    int
	extendStyle bool
}

func (c croppedLines) Render(width, height int) *term.Buffer {
	bb := term.NewBufferBuilder(width)
	leftSpacing := ui.NewText(strings.Repeat(" ", c.padding))
	rightSpacing := ui.NewText(strings.Repeat(" ", width-c.padding))
	for i, line := range c.lines {
		if i > 0 {
			bb.Newline()
		}

		selected := c.selectFrom <= i && i < c.selectTo
		extendStyle := c.extendStyle && len(line) > 0

		left := leftSpacing.Clone()
		if extendStyle {
			left[0].Style = line[0].Style
		}
		acc := left.ConcatText(line.TrimWcwidth(width - 2*c.padding))
		if extendStyle || selected {
			right := rightSpacing.Clone()
			if extendStyle {
				right[0].Style = line[len(line)-1].Style
			}
			acc = acc.ConcatText(right).TrimWcwidth(width)
		}
		if selected {
			acc = ui.Transform(acc, styleForSelected)
		}

		bb.WriteStyled(acc)
	}
	return bb.Buffer()
}

func (w *widget) Handle(event term.Event) bool {
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

func (w *widget) CopyState() State {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

func (w *widget) Reset(it Items, selected int) {
	w.mutate(func(s *State) { *s = State{Items: it, Selected: selected} })
	if 0 <= selected && selected < it.Len() {
		w.OnSelect(it, selected)
	}
}

func (w *widget) Select(f func(State) int) {
	var it Items
	var oldSelected, selected int
	w.mutate(func(s *State) {
		oldSelected, it = s.Selected, s.Items
		selected = f(*s)
		s.Selected = selected
	})
	if selected != oldSelected && 0 <= selected && selected < it.Len() {
		w.OnSelect(it, selected)
	}
}

// Prev moves the selection to the previous item, or does nothing if the
// first item is currently selected. It is a suitable as an argument to
// Widget.Select.
func Prev(s State) int {
	return fixIndex(s.Selected-1, s.Items.Len())
}

// PrevPage moves the selection to the item one page before. It is only
// meaningful in vertical layout and suitable as an argument to Widget.Select.
//
// TODO(xiaq): This does not correctly with multi-line items.
func PrevPage(s State) int {
	return fixIndex(s.Selected-s.Height, s.Items.Len())
}

// Next moves the selection to the previous item, or does nothing if the
// last item is currently selected. It is a suitable as an argument to
// Widget.Select.
func Next(s State) int {
	return fixIndex(s.Selected+1, s.Items.Len())
}

// NextPage moves the selection to the item one page after. It is only
// meaningful in vertical layout and suitable as an argument to Widget.Select.
//
// TODO(xiaq): This does not correctly with multi-line items.
func NextPage(s State) int {
	return fixIndex(s.Selected+s.Height, s.Items.Len())
}

// PrevWrap moves the selection to the previous item, or to the last item if
// the first item is currently selected. It is a suitable as an argument to
// Widget.Select.
func PrevWrap(s State) int {
	selected, n := s.Selected, s.Items.Len()
	switch {
	case selected >= n:
		return n - 1
	case selected <= 0:
		return n - 1
	default:
		return selected - 1
	}
}

// NextWrap moves the selection to the previous item, or to the first item
// if the last item is currently selected. It is a suitable as an argument to
// Widget.Select.
func NextWrap(s State) int {
	selected, n := s.Selected, s.Items.Len()
	switch {
	case selected >= n-1:
		return 0
	case selected < 0:
		return 0
	default:
		return selected + 1
	}
}

// Left moves the selection to the item to the left. It is only meaningful in
// horizontal layout and suitable as an argument to Widget.Select.
func Left(s State) int {
	return horizontal(s.Selected, s.Items.Len(), -s.Height)
}

// Right moves the selection to the item to the right. It is only meaningful in
// horizontal layout and suitable as an argument to Widget.Select.
func Right(s State) int {
	return horizontal(s.Selected, s.Items.Len(), s.Height)
}

func horizontal(selected, n, d int) int {
	selected = fixIndex(selected, n)
	newSelected := selected + d
	if newSelected < 0 || newSelected >= n {
		return selected
	}
	return newSelected
}

func fixIndex(i, n int) int {
	switch {
	case i < 0:
		return 0
	case i >= n:
		return n - 1
	default:
		return i
	}
}

func (w *widget) Accept() {
	state := w.CopyState()
	w.OnAccept(state.Items, state.Selected)
}

func (w *widget) mutate(f func(s *State)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}
