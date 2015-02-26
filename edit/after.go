package edit

import "time"

// After is like time.After, except that it interpretes negative Duration as
// "forever".
func After(d time.Duration) <-chan time.Time {
	if d < 0 {
		return make(chan time.Time)
	}
	return time.After(d)
}
