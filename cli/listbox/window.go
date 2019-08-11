package listbox

// The number of lines the listing mode keeps between the current selected item
// and the top and bottom edges of the window, unless the available height is
// too small or if the selected item is near the top or bottom of the list.
var respectDistance = 2

// Determines the index of the first item to show in listing.
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
func findWindow(itemer Itemer, n, selected, lastFirst, height int) (first, crop int) {
	selectedHeight := itemer.Item(selected).CountLines()

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
		useDown += itemer.Item(i).CountLines()
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
		useUp += itemer.Item(i).CountLines()
		if useUp >= budgetUp {
			return i, useUp - budgetUp
		}
		if i <= lastFirst && useUp >= respectDistance && useUp+useDown >= budget {
			return i, 0
		}
	}
	return 0, 0
}
