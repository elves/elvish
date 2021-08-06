// Package testutil contains common test utilities.
package testutil

// Cleanuper wraps the Cleanup method. It is a subset of testing.TB, thus
// satisfied by *testing.T and *testing.B.
type Cleanuper interface {
	Cleanup(func())
}
