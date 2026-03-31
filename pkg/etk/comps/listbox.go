package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/mtd"
	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

// Interface for the items state of a listbox.
var (
	// ListItemsLen returns the number of items.
	ListItemsLen = mtd.New[func(fm *eval.Frame, x any) (int, error)]("Len")
	// ListItemsGet accesses the underlying item.
	ListItemsGet = mtd.New[func(fm *eval.Frame, x any, i int) (any, error)]("Get")
	// ListItemsShow renders the item at the given zero-based index,
	// also returning the "area styling" for the item,
	// which is applied to the entire area occupied by the item in the listbox.
	ListItemsShow = mtd.New[func(fm *eval.Frame, x any, i int) (ui.Text, ui.Styling, error)]("Show")
)

func CallListItemsLen(c etk.Context, x any) int {
	n, err := ListItemsLen.Call(c.Frame(), x)
	if err != nil {
		return 0
	}
	return n
}

func CallListItemsGet(c etk.Context, x any, i int) any {
	item, err := ListItemsGet.Call(c.Frame(), x, i)
	if err != nil {
		return nil
	}
	return item
}

func CallListItemsShow(c etk.Context, x any, i int) (ui.Text, ui.Styling) {
	t, st, err := ListItemsShow.Call(c.Frame(), x, i)
	if err != nil {
		return ui.T("???"), ui.FgRed
	}
	return t, st
}

type StringItems []string

// MakeStringItems returns a value to be used for the items state of a listbox,
// backed by a slice of strings.
func MakeStringItems(items ...string) StringItems       { return StringItems(items) }
func (si StringItems) Len() int                         { return len(si) }
func (si StringItems) Get(i int) any                    { return si[i] }
func (si StringItems) Show(i int) (ui.Text, ui.Styling) { return ui.T(si[i]), ui.Nop }

// ListBox shows a list of items and supports choosing one of them.
//
// State variables:
//
//   - items: a list of items
//   - selected: an int storing the index of the selected item
func ListBox(c etk.Context) (etk.View, etk.React) {
	// Essential state variables.
	itemsVar := etk.State(c, "items", any(nil))
	selectedVar := etk.State(c, "selected", 0)
	// Layout configuration variables.
	multiColumnVar := etk.State(c, "multi-column", false)
	leftPaddingVar := etk.State(c, "left-padding", 0)
	rightPaddingVar := etk.State(c, "right-padding", 0)
	// Internal UI state (see also comment in listBoxView).
	firstVar := etk.State(c, "-first", 0)
	contentHeightVar := etk.State(c, "-content-height", 0)

	view := &listBoxView{
		c, itemsVar.Get(), selectedVar.Get(),
		multiColumnVar.Get(), leftPaddingVar.Get(), rightPaddingVar.Get(),
		firstVar, contentHeightVar}
	return view,
		c.Binding(func(e term.Event) etk.Reaction {
			selected := selectedVar.Get()
			items := itemsVar.Get()
			n := CallListItemsLen(c, items)
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
	c            etk.Context
	items        any
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
	if v.items == nil || CallListItemsLen(v.c, v.items) == 0 {
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
	first, firstCrop := singleColumnWindow(v.c, v.items, v.selected, v.first.Get(), height)
	v.first.Set(first)

	lv := linesView{
		LeftPadding: v.leftPadding, RightPadding: v.rightPadding}
	n := CallListItemsLen(v.c, v.items)
	var i int
	for i = first; i < n && len(lv.Lines) < height; i++ {
		text, areaStyling := CallListItemsShow(v.c, v.items, i)
		if i == v.selected {
			lv.DotAtLine = len(lv.Lines)
			areaStyling = ui.Stylings(areaStyling, ui.Inverse)
		}

		lines := text.SplitByRune('\n')
		if i == first {
			lines = lines[firstCrop:]
		}
		for _, line := range lines {
			if len(lv.Lines) == height {
				break
			}
			lv.Lines = append(lv.Lines, line)
			lv.LineStylings = append(lv.LineStylings, areaStyling)
		}
	}
	if first == 0 && i == n && firstCrop == 0 && len(lv.Lines) < height {
		return lv.Render(width, height)
	}
	box := etk.Box("content* scrollbar=",
		&lv, etk.ScrollBarView{Total: n, Low: first, High: i})
	return box.Render(width, height)
}

func (w *listBoxView) renderMultiColumn(width, height int) *term.Buffer {
	// TODO: Make padding customizable
	first, colHeight, _ := multiColumnWindow(
		w.c, w.items, w.selected, w.first.Get(), w.leftPadding+w.rightPadding, width, height)
	w.first.Set(first)
	w.contentHeight.Set(colHeight)

	items, selected, first := w.items, w.selected, w.first.Get()
	n := CallListItemsLen(w.c, items)

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
			text, areaStyling := CallListItemsShow(w.c, items, j)
			if j == selected {
				col.DotAtLine = len(col.Lines)
				areaStyling = ui.Stylings(areaStyling, ui.Inverse)
			}

			// TODO: Complain about multi-line items more loudly.
			col.Lines = append(col.Lines, text.SplitByRune('\n')[0])
			col.LineStylings = append(col.LineStylings, areaStyling)
		}

		colWidth := maxWidth(w.c, items, w.leftPadding+w.rightPadding, i, i+colHeight)
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
