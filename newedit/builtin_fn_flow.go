package newedit

import (
	"github.com/elves/elvish/newedit/editutil"
	"github.com/elves/elvish/newedit/types"
)

//elvdoc:fn exit-binding
//
// Exits the current binding handler. Internally, this works by raising a
// special exception.

func exitBinding() error {
	return editutil.ActionError(types.NoAction)
}

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

func commitCode() error {
	return editutil.ActionError(types.CommitCode)
}

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func commitEOF() error {
	return editutil.ActionError(types.CommitEOF)
}
