package listbox

import "github.com/elves/elvish/util"

// The number of lines the listing mode keeps between the current selected item
// and the top and bottom edges of the window, unless the available height is
// too small or if the selected item is near the top or bottom of the list.
var respectDistance = 2

// Determines the index of the first item to show in vertical layout.
//
// This function does not return the full window, but just the first item to
// show, and how many initial lines to crop. The window determined by this
// algorithm has the following properties:
//
// * It always includes the selected item.
//
// * The combined height of all the entries in the window is equal to
//   min(height, combined height of all entries).
//
// * There are at least respectDistance rows above the first row of the selected
//   item, as well as that many rows below the last row of the selected item,
//   unless the height is too small.
//
// * Among all values satisfying the above conditions, the value of first is
//   the one closest to lastFirst.
func getVertialWindow(state State, height int) (first, crop int) {
	items, selected, lastFirst := state.Items, state.Selected, state.First
	n := items.Len()
	selectedHeight := items.Show(selected).CountLines()

	if height <= selectedHeight {
		// The height is not big enough (or just big enough) to fit the selected
		// item. Fit as much as the selected item as we can.
		return selected, 0
	}

	// Determine the minimum amount of space required for the downward direction.
	budget := height - selectedHeight
	var needDown int
	if budget >= 2*respectDistance {
		// If we can afford maintaining the respect distance on both sides, then
		// the minimum amount of space required is the respect distance.
		needDown = respectDistance
	} else {
		// Otherwise we split the available space by half. The downward (no pun
		// intended) rounding here is an arbitrary choice.
		needDown = budget / 2
	}
	// Calculate how much of the budget the downward direction can use. This is
	// used to 1) potentially shrink needDown 2) decide how much to expand
	// upward later.
	useDown := 0
	for i := selected + 1; i < n; i++ {
		useDown += items.Show(i).CountLines()
		if useDown >= budget {
			break
		}
	}
	if needDown > useDown {
		// We reached the last item without using all of needDown. That means we
		// don't need so much in the downward direction.
		needDown = useDown
	}

	// The maximum amount of space we can use in the upward direction is the
	// entire budget minus the minimum amount of space we need in the downward
	// direction.
	budgetUp := budget - needDown

	useUp := 0
	// Extend upwards until any of the following becomes true:
	//
	// * We have exhausted budgetUp;
	//
	// * We have reached item 0;
	//
	// * We have reached or passed lastFirst, satisfied the upward respect
	//   distance, and will be able to use up the entire budget when expanding
	//   downwards later.
	for i := selected - 1; i >= 0; i-- {
		useUp += items.Show(i).CountLines()
		if useUp >= budgetUp {
			return i, useUp - budgetUp
		}
		if i <= lastFirst && useUp >= respectDistance && useUp+useDown >= budget {
			return i, 0
		}
	}
	return 0, 0
}

// Determines the window to show in horizontal layout. It returns the first item
// to show, and whether all items can fit in the window.
func getHorizontalWindow(state State, width, height int) (int, bool) {
	items := state.Items
	n := items.Len()
	if n <= height {
		// All items can fit.
		return 0, true
	}
	// Lower bound of number of items that can fit in a row.
	perRow := (width + colGap) / (maxWidth(items, 0, n) + colGap)
	if perRow == 0 {
		// We trim items that are too wide, so there is at least one item per row.
		perRow = 1
	}
	if height*perRow >= n {
		// All items can fit.
		return 0, true
	}
	// Reduce the amount of available height by one because the last row will be
	// reserved for the scrollbar.
	height--
	selected, lastFirst := state.Selected, state.First
	// Start with the column containing the selected item, move left until
	// either the width is exhausted, or lastFirst has been reached.
	first := selected / height * height
	usedWidth := maxWidth(items, first, first+height)
	for ; first > lastFirst; first -= height {
		usedWidth += maxWidth(items, first-height, first) + colGap
		if usedWidth > width {
			break
		}
	}
	return first, false
}

func maxWidth(items Items, low, high int) int {
	n := items.Len()
	width := 0
	for i := low; i < high && i < n; i++ {
		w := 0
		for _, seg := range items.Show(i) {
			w += util.Wcswidth(seg.Text)
		}
		if width < w {
			width = w
		}
	}
	return width
}
