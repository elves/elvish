package edit

import "github.com/elves/elvish/edit/edtypes"

// Aliases for variables and functions from the edtypes package.
// NOTE: We cannot use type aliases before we drop support for Go 1.8.
var (
	noAction     = edtypes.NoAction
	reprocessKey = edtypes.ReprocessKey
	commitLine   = edtypes.CommitLine
	commitEOF    = edtypes.CommitEOF

	emptyBindingMap = edtypes.EmptyBindingMap
)
