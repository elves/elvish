package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
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

// StyleLiner is an optional interface that [ListItems] can implement.
type StyleLiner interface {
	// StyleLine returns a "line styling" for item i, which gets applied to
	// whole lines occupied by the item, including empty spaces.
	StyleLine(i int) ui.Styling
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
	leftPaddingVar := etk.State(c, "left-padding", 0)
	rightPaddingVar := etk.State(c, "right-padding", 0)
	// Internal UI state (see also comment in listBoxView).
	firstVar := etk.State(c, "-first", 0)
	contentHeightVar := etk.State(c, "-content-height", 0)

	view := &listBoxView{
		itemsVar.Get(), selectedVar.Get(),
		multiColumnVar.Get(), leftPaddingVar.Get(), rightPaddingVar.Get(),
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
	items        ListItems
	selected     int
	multiColumn  bool
	leftPadding  int
	rightPadding int
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

func (v *listBoxView) renderSingleColumn(width, height int) *term.Buffer {
	first, firstCrop := getVerticalWindow(v.items, v.selected, v.first.Get(), height)
	v.first.Set(first)

	lv := linesView{
		LeftPadding: v.leftPadding, RightPadding: v.rightPadding}
	n := v.items.Len()
	var i int
	for i = first; i < n && len(lv.Lines) < height; i++ {
		text := v.items.Show(i)
		lineStyling := ui.Nop
		if styleLiner, ok := v.items.(StyleLiner); ok {
			lineStyling = styleLiner.StyleLine(i)
		}
		if i == v.selected {
			lv.DotAtLine = len(lv.Lines)
			lineStyling = ui.Stylings(lineStyling, ui.Inverse)
		}

		lines := text.SplitByRune('\n')
		if i == first {
			lines = lines[firstCrop:]
		}
		for _, line := range lines {
			lv.Lines = append(lv.Lines, line)
			lv.LineStylings = append(lv.LineStylings, lineStyling)
		}
	}
	if first == 0 && i == n && firstCrop == 0 && len(lv.Lines) < height {
		return lv.Render(width, height)
	}
	box := etk.Box("content* scrollbar=",
		&lv, etk.ScrollBarView{Total: n, Low: first, High: i})
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
		col := linesView{
			LeftPadding: w.leftPadding, RightPadding: w.rightPadding}

		// Render the column starting from i.
		for j := i; j < i+colHeight && j < n; j++ {
			last = j
			text := items.Show(j)
			lineStyling := ui.Nop
			if styleLiner, ok := w.items.(StyleLiner); ok {
				lineStyling = styleLiner.StyleLine(i)
			}
			if j == selected {
				col.DotAtLine = len(col.Lines)
				lineStyling = ui.Stylings(lineStyling, ui.Inverse)
			}

			// TODO: Complain about multi-line items more loudly.
			col.Lines = append(col.Lines, text.SplitByRune('\n')[0])
			col.LineStylings = append(col.LineStylings, lineStyling)
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

// A specialized line-oriented View for ListBox.
//
// Ideally we would like to use etk.TextView. However, etk.TextView has a
// text-oriented API. ListBox needs support for line padding and line styling,
// which are quite awkward to add to TextView. This type has a line-oriented
// API and makes these two features easier to implement.
//
// The downside of this implementation is that linesView doesn't support
// wrapping; each line is cropped.
//
// The user of linesView is responsible for ensuring that:
//
//   - len(v.Lines) < height
//   - len(v.Lines) == len(v.LineStylings)
type linesView struct {
	LeftPadding  int
	RightPadding int
	Lines        []ui.Text
	LineStylings []ui.Styling
	DotAtLine    int
}

func (v *linesView) Render(width, height int) *term.Buffer {
	buf := term.Buffer{Width: width, Dot: term.Pos{Line: v.DotAtLine, Col: 0}}
	leftPadding, rightPadding := v.LeftPadding, v.RightPadding
	if leftPadding+rightPadding >= width {
		leftPadding, rightPadding = 0, 0
	}
	for i, line := range v.Lines {
		lineStyling := v.LineStylings[i]
		paddingCell := term.Cell{
			Text: " ", Style: ui.ApplyStyling(ui.Style{}, lineStyling).SGR()}

		var bufLine []term.Cell
		for range leftPadding {
			bufLine = append(bufLine, paddingCell)
		}
		col := leftPadding

	renderLineContent:
		for _, seg := range line {
			segSGR := ui.ApplyStyling(seg.Style, lineStyling).SGR()
			for _, r := range seg.Text {
				cell := etk.PrintCell(r, segSGR)
				cellWidth := wcwidth.Of(cell.Text)
				if col+cellWidth+rightPadding > width {
					break renderLineContent
				}
				bufLine = append(bufLine, cell)
				col += cellWidth
			}
		}

		for range width - col {
			bufLine = append(bufLine, paddingCell)
		}
		buf.Lines = append(buf.Lines, bufLine)
	}
	return &buf
}
