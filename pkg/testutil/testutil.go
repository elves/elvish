// Package testutil contains common test utilities.
package testutil

// Cleanuper wraps the Cleanup method. It is a subset of [testing.TB], thus
// satisfied by [*testing.T] and [*testing.B].
type Cleanuper interface {
	Cleanup(func())
}

// Skipper wraps the Skipf method. It is a subset of [testing.TB], thus
// satisfied by [*testing.T] and [*testing.B].
type Skipper interface {
	Skipf(format string, args ...any)
}
