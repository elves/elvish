// Package storedefs contains definitions used by the store package.
package storedefs

import "errors"

// NoBlacklist is an empty blacklist, to be used in GetDirs.
var NoBlacklist = map[string]struct{}{}

// ErrNoMatchingCmd is the error returned when a LastCmd or FirstCmd query
// completes with no result.
var ErrNoMatchingCmd = errors.New("no matching command line")

// Dir is an entry in the directory history.
type Dir struct {
	Path  string
	Score float64
}
