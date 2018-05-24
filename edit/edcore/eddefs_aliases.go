package edcore

import "github.com/elves/elvish/edit/eddefs"

// Aliases for variables and functions from the eddefs package.
// NOTE: We cannot use type aliases before we drop support for Go 1.8.
var (
	noAction     = eddefs.NoAction
	reprocessKey = eddefs.ReprocessKey
	commitLine   = eddefs.CommitLine
	commitEOF    = eddefs.CommitEOF

	emptyBindingMap = eddefs.EmptyBindingMap
)
