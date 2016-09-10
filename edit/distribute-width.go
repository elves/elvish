package edit

func distributeWidths(w int, weights []float64, actual []int) []int {
	n := len(weights)
	widths := make([]int, n)
	done := make([]bool, n)

	for {
		wsum := 0.0
		for i := 0; i < n; i++ {
			if done[i] {
				continue
			}
			wsum += weights[i]
		}
		// Widths allocated away
		allocated := 0
		for i := 0; i < n; i++ {
			if done[i] {
				continue
			}
			allowed := int(float64(w) * weights[i] / wsum)
			if actual[i] <= allowed {
				// Actual width fit in allowed width; allocate
				widths[i] = actual[i]
				allocated += actual[i]
				done[i] = true
			}
		}
		if allocated == 0 {
			// Use allowed width for all remaining columns
			for i := 0; i < n; i++ {
				if done[i] {
					continue
				}
				allowed := int(float64(w) * weights[i] / wsum)
				widths[i] = allowed
				w -= allowed
				done[i] = true
			}
			break
		}
	}
	Logger.Printf("distribute(%d, %v, %v) -> %v", w, weights, actual, widths)
	return widths
}
