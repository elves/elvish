package edcore

// getNavWidths calculates the widths for the three (parent, current and
// preview) columns in the navigation mode. It takes the available width, full
// width required to display the current and preview columns, and returns
// suitable widths for the columns.
//
// The parent column always gets 1/6 of the total width. The current and preview
// columns initially get 1/2 of the remaining width each, but if one of them
// does not have enough widths and another has some spare width, the amount of
// spare width or the needed width (whichever is smaller) is donated from the
// latter to the former.
func getNavWidths(total, currentFull, previewFull int) (int, int, int) {
	parent := total / 6

	remain := total - parent
	current := remain / 2
	preview := remain - current

	if current < currentFull && preview > previewFull {
		donate := min(currentFull-current, preview-previewFull)
		current += donate
		preview -= donate
	} else if preview < previewFull && current > currentFull {
		donate := min(previewFull-preview, current-currentFull)
		preview += donate
		current -= donate
	}
	return parent, current, preview
}
