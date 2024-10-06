package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

// ListItems is an interface for accessing multiple items.
type ListItems interface {
	// Len returns the number of items.
	Len() int
	// Get accesses the underlying item.
	Get(i int) any
	// Show renders the item at the given zero-based index.
	Show(i int) ui.Text
}

type stringItems []string

// StringItems returns a [ListItems] backed up a slice of strings.
func StringItems(items ...string) ListItems { return stringItems(items) }
func (si stringItems) Len() int             { return len(si) }
func (si stringItems) Get(i int) any        { return si[i] }
func (si stringItems) Show(i int) ui.Text   { return ui.T(si[i]) }

func ListBox(c etk.Context) (etk.View, etk.React) {
	// Essential state variables.
	itemsVar := etk.State(c, "items", ListItems(nil))
	selectedVar := etk.State(c, "selected", 0)
	// Layout configuration variables.
	multiColumnVar := etk.State(c, "multi-column", false)
	// Internal UI state (see also comment in listBoxView).
	firstVar := etk.State(c, "-first", 0)
	contentHeightVar := etk.State(c, "-content-height", 0)

	view := &listBoxView{
		itemsVar.Get(), selectedVar.Get(), multiColumnVar.Get(),
		firstVar, contentHeightVar}
	return view,
		c.Binding(func(e term.Event) etk.Reaction {
			selected := selectedVar.Get()
			items := itemsVar.Get()
			n := items.Len()
			switch e {
			case term.K(ui.Up):
				if selected-1 >= 0 {
					selectedVar.Set(selected - 1)
					return etk.Consumed
				}
			case term.K(ui.Down):
				if selected+1 < n {
					selectedVar.Set(selected + 1)
					return etk.Consumed
				}
			case term.K(ui.Tab, ui.Shift):
				selectedVar.Set((selected + n - 1) % n)
				return etk.Consumed
			case term.K(ui.Tab):
				selectedVar.Set((selected + 1) % n)
				return etk.Consumed
			}
			if multiColumnVar.Get() {
				contentHeight := contentHeightVar.Get()
				switch e {
				case term.K(ui.Left):
					if selected-contentHeight >= 0 {
						selectedVar.Set(selected - contentHeight)
						return etk.Consumed
					}
				case term.K(ui.Right):
					if selected+contentHeight < n {
						selectedVar.Set(selected + contentHeight)
						return etk.Consumed
					}
				}
			}
			return etk.Unused
		})
}

type listBoxView struct {
	items       ListItems
	selected    int
	multiColumn bool
	// The first element that was shown last time.
	//
	// Used to provide some continuity in the UI when the terminal size has
	// changed or when the listbox has been scrolled.
	first etk.StateVar[int]
	// Height of the listbox, excluding horizontal scrollbar when using the
	// horizontal layout (hence content height). Stored in the state for
	// commands to move the cursor by page (for vertical layout) or column (for
	// horizontal layout).
	contentHeight etk.StateVar[int]
}

func (v *listBoxView) Render(width, height int) *term.Buffer {
	if v.items == nil || v.items.Len() == 0 {
		v.first.Set(0)
		v.contentHeight.Set(1)
		// TODO: Respect height; make placeholder customization
		return term.NewBufferBuilder(width).Write("(no item)").Buffer()
	}

	if v.multiColumn {
		return v.renderMultiColumn(width, height)
	} else {
		return v.renderSingleColumn(width, height)
	}
}

func (w *listBoxView) renderSingleColumn(width, height int) *term.Buffer {
	first, firstCrop := getVerticalWindow(w.items, w.selected, w.first.Get(), height)
	w.first.Set(first)

	v := etk.TextView{Wrap: etk.NoWrap}
	lines := 0
	n := w.items.Len()
	var i int
	for i = first; i < n && lines < height; i++ {
		if i > first {
			v.Spans = append(v.Spans, ui.T("\n"))
		}

		text := w.items.Show(i)
		if i == w.selected {
			v.DotBefore = len(v.Spans)
			text = ui.StyleText(text, ui.Inverse)
		}

		if i == first {
			keptLines := text.SplitByRune('\n')[firstCrop:]
			for i, line := range keptLines {
				if i > 0 {
					v.Spans = append(v.Spans, ui.T("\n"))
				}
				v.Spans = append(v.Spans, line)
			}
			lines += len(keptLines)
		} else {
			v.Spans = append(v.Spans, text)
			lines += text.CountLines()
		}
	}
	if first == 0 && i == n && firstCrop == 0 && lines < height {
		return v.Render(width, height)
	}
	box := etk.Box("content* scrollbar=",
		v, etk.ScrollBarView{Total: n, Low: first, High: i})
	return box.Render(width, height)
}

const padding = 1

func (w *listBoxView) renderMultiColumn(width, height int) *term.Buffer {
	// TODO: Make padding customizable
	first, colHeight, _ := getHorizontalWindow(w.items, w.selected, w.first.Get(), padding, width, height)
	w.first.Set(first)
	w.contentHeight.Set(colHeight)

	items, selected, first := w.items, w.selected, w.first.Get()
	n := items.Len()

	buf := &term.Buffer{}
	remainedWidth := width
	hasCropped := false
	last := first
	for i := first; i < n; i += colHeight {
		col := etk.TextView{Wrap: etk.NoWrap}
		// Render the column starting from i.
		for j := i; j < i+colHeight && j < n; j++ {
			last = j
			if j > i {
				col.Spans = append(col.Spans, ui.T("\n"))
			}
			text := items.Show(j)
			if j == selected {
				text = ui.StyleText(text, ui.Inverse)
				col.DotBefore = len(col.Spans)
			}
			col.Spans = append(col.Spans, text)
		}

		colWidth := maxWidth(items, padding, i, i+colHeight)
		if colWidth > remainedWidth {
			colWidth = remainedWidth
			hasCropped = true
		}

		buf.ExtendRight(
			col.Render(colWidth, colHeight),
			i <= selected && selected < i+colHeight)

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
		scrollbar := etk.ScrollBarView{
			Horizontal: true, Total: n, Low: first, High: last + 1}
		buf.ExtendDown(scrollbar.Render(width, 1), false)
	}
	return buf
}
