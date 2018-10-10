package util

// Feed calls the function with given values, breaking earlier if the function
// returns false.
func Feed(f func(interface{}) bool, values ...interface{}) {
	for _, value := range values {
		if !f(value) {
			break
		}
	}
}
