package parse

import (
	"strconv"
)

// Atou is basically shorthand for strconv.ParseUint(s, 10, 0) but returns the
// first argument as uintptr. Useful for parsing fd.
func Atou(s string) (uintptr, error) {
	u, err := strconv.ParseUint(s, 10, 0)
	return uintptr(u), err
}
