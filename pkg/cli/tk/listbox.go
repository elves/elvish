package tk

import (
	"strings"
	"sync"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/ui"
)

// ListBox is a list for displaying and selecting from a list of items.
type ListBox interface {
	Widget
	// CopyState returns a copy of the state.
	CopyState() ListBoxState
	// Reset resets the state of the widget with the given items and index of
	// the selected item. It triggers the OnSelect callback if the index is
	// valid.
	Reset(it Items, selected int)
	// Select changes the selection by calling f with the current state, and
	// using the return value as the new selection index. It triggers the
	// OnSelect callback if the selected index has changed and is valid.
	Select(f func(ListBoxState) int)
	// Accept accepts the currently selected item.
	Accept()
}

// ListBoxSpec specifies the configuration and initial state for ListBox.
type ListBoxSpec struct {
	// Key bindings.
	Bindings Bindings
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

	// State. When used in [NewListBox], this field specifies the initial state.
	State ListBoxState
}

type listBox struct {
	// Mutex for synchronizing access to the state.
	StateMutex sync.RWMutex
	// Configuration and state.
	ListBoxSpec
}

// NewListBox creates a new ListBox from the given spec.
func NewListBox(spec ListBoxSpec) ListBox {
	if spec.Bindings == nil {
		spec.Bindings = DummyBindings{}
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
	return &listBox{ListBoxSpec: spec}
}

var stylingForSelected = ui.Inverse

func (w *listBox) Render(width, height int) *term.Buffer {
	if w.Horizontal {
		return w.renderHorizontal(width, height)
	}
	return w.renderVertical(width, height)
}

func (w *listBox) MaxHeight(width, height int) int {
	s := w.CopyState()
	if s.Items == nil || s.Items.Len() == 0 {
		return 0
	}
	if w.Horizontal {
		_, h, scrollbar := getHorizontalWindow(s, w.Padding, width, height)
		if scrollbar {
			return h + 1
		}
		return h
	}
	h := 0
	for i := 0; i < s.Items.Len(); i++ {
		h += s.Items.Show(i).CountLines()
		if h >= height {
			return height
		}
	}
	return h
}

const listBoxColGap = 2

func (w *listBox) renderHorizontal(width, height int) *term.Buffer {
	var state ListBoxState
	var colHeight int
	w.mutate(func(s *ListBoxState) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, s.ContentHeight, _ = getHorizontalWindow(*s, w.Padding, width, height)
			colHeight = s.ContentHeight
		}
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return Label{Content: w.Placeholder}.Render(width, height)
	}

	items, selected, first := state.Items, state.Selected, state.First
	n := items.Len()

	buf := term.NewBuffer(0)
	remainedWidth := width
	hasCropped := false
	last := first
	for i := first; i < n; i += colHeight {
		selectedRow := -1
		// Render the column starting from i.
		col := make([]ui.Text, 0, colHeight)
		for j := i; j < i+colHeight && j < n; j++ {
			last = j
			item := items.Show(j)
			if j == selected {
				selectedRow = j - i
			}
			col = append(col, item)
		}

		colWidth := maxWidth(items, w.Padding, i, i+colHeight)
		if colWidth > remainedWidth {
			colWidth = remainedWidth
			hasCropped = true
		}

		colBuf := croppedLines{
			lines: col, padding: w.Padding,
			selectFrom: selectedRow, selectTo: selectedRow + 1,
			extendStyle: w.ExtendStyle}.Render(colWidth, colHeight)
		buf.ExtendRight(colBuf)

		remainedWidth -= colWidth
		if remainedWidth <= listBoxColGap {
			break
		}
		remainedWidth -= listBoxColGap
		buf.Width += listBoxColGap
	}
	// We may not have used all the width required; force buffer width.
	buf.Width = width
	if colHeight < height && (first != 0 || last != n-1 || hasCropped) {
		scrollbar := HScrollbar{Total: n, Low: first, High: last + 1}
		buf.Extend(scrollbar.Render(width, 1), false)
	}
	return buf
}

func (w *listBox) renderVertical(width, height int) *term.Buffer {
	var state ListBoxState
	var firstCrop int
	w.mutate(func(s *ListBoxState) {
		if s.Items == nil || s.Items.Len() == 0 {
			s.First = 0
		} else {
			s.First, firstCrop = getVerticalWindow(*s, height)
		}
		s.ContentHeight = height
		state = *s
	})

	if state.Items == nil || state.Items.Len() == 0 {
		return Label{Content: w.Placeholder}.Render(width, height)
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

	var rd Renderer = croppedLines{
		lines: allLines, padding: w.Padding,
		selectFrom: selectFrom, selectTo: selectTo, extendStyle: w.ExtendStyle}
	if first > 0 || i < n || hasCropped {
		rd = VScrollbarContainer{
			Content:   rd,
			Scrollbar: VScrollbar{Total: n, Low: first, High: i},
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
	leftSpacing := ui.T(strings.Repeat(" ", c.padding))
	rightSpacing := ui.T(strings.Repeat(" ", width-c.padding))
	for i, line := range c.lines {
		if i > 0 {
			bb.Newline()
		}

		selected := c.selectFrom <= i && i < c.selectTo
		extendStyle := c.extendStyle && len(line) > 0

		left := leftSpacing.Clone()
		if extendStyle && len(left) > 0 {
			left[0].Style = line[0].Style
		}
		acc := ui.Concat(left, line.TrimWcwidth(width-2*c.padding))
		if extendStyle || selected {
			right := rightSpacing.Clone()
			if extendStyle {
				right[0].Style = line[len(line)-1].Style
			}
			acc = ui.Concat(acc, right).TrimWcwidth(width)
		}
		if selected {
			acc = ui.StyleText(acc, stylingForSelected)
		}

		bb.WriteStyled(acc)
	}
	return bb.Buffer()
}

func (w *listBox) Handle(event term.Event) bool {
	if w.Bindings.Handle(w, event) {
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

func (w *listBox) CopyState() ListBoxState {
	w.StateMutex.RLock()
	defer w.StateMutex.RUnlock()
	return w.State
}

func (w *listBox) Reset(it Items, selected int) {
	w.mutate(func(s *ListBoxState) { *s = ListBoxState{Items: it, Selected: selected} })
	if 0 <= selected && selected < it.Len() {
		w.OnSelect(it, selected)
	}
}

func (w *listBox) Select(f func(ListBoxState) int) {
	var it Items
	var oldSelected, selected int
	w.mutate(func(s *ListBoxState) {
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
// [ListBox.Select].
func Prev(s ListBoxState) int {
	return fixIndex(s.Selected-1, s.Items.Len())
}

// PrevPage moves the selection to the item one page before. It is only
// meaningful in vertical layout and suitable as an argument to
// [ListBox.Select].
//
// TODO(xiaq): This does not correctly with multi-line items.
func PrevPage(s ListBoxState) int {
	return fixIndex(s.Selected-s.ContentHeight, s.Items.Len())
}

// Next moves the selection to the previous item, or does nothing if the
// last item is currently selected. It is a suitable as an argument to
// [ListBox.Select].
func Next(s ListBoxState) int {
	return fixIndex(s.Selected+1, s.Items.Len())
}

// NextPage moves the selection to the item one page after. It is only
// meaningful in vertical layout and suitable as an argument to
// [ListBox.Select].
//
// TODO(xiaq): This does not correctly with multi-line items.
func NextPage(s ListBoxState) int {
	return fixIndex(s.Selected+s.ContentHeight, s.Items.Len())
}

// PrevWrap moves the selection to the previous item, or to the last item if
// the first item is currently selected. It is a suitable as an argument to
// [ListBox.Select].
func PrevWrap(s ListBoxState) int {
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
// [ListBox.Select].
func NextWrap(s ListBoxState) int {
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
// horizontal layout and suitable as an argument to [ListBox.Select].
func Left(s ListBoxState) int {
	return horizontal(s.Selected, s.Items.Len(), -s.ContentHeight)
}

// Right moves the selection to the item to the right. It is only meaningful in
// horizontal layout and suitable as an argument to [ListBox.Select].
func Right(s ListBoxState) int {
	return horizontal(s.Selected, s.Items.Len(), s.ContentHeight)
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

func (w *listBox) Accept() {
	state := w.CopyState()
	if 0 <= state.Selected && state.Selected < state.Items.Len() {
		w.OnAccept(state.Items, state.Selected)
	}
}

func (w *listBox) mutate(f func(s *ListBoxState)) {
	w.StateMutex.Lock()
	defer w.StateMutex.Unlock()
	f(&w.State)
}
