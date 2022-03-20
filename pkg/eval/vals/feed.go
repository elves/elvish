package vals

// Feed calls the function with given values, breaking earlier if the function
// returns false.
func Feed(f func(any) bool, values ...any) {
	for _, value := range values {
		if !f(value) {
			break
		}
	}
}
