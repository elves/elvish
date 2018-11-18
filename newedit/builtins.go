package newedit

import (
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/newedit/editutil"
	"github.com/elves/elvish/newedit/types"
)

//elvish:doc-fn binding-map
//
// Converts a normal map into a binding map.

var makeBindingMap = eddefs.MakeBindingMap

//elvish:doc-fn exit-binding
//
// Exits the current binding handler. Internally, this works by raising a
// special exception.

func exitBinding() error {
	return editutil.ActionError(types.NoAction)
}

//elvish:doc-fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

func commitCode() error {
	return editutil.ActionError(types.CommitCode)
}

//elvish:doc-fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func commitEOF() error {
	return editutil.ActionError(types.CommitEOF)
}
