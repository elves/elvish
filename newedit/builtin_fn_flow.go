package newedit

import (
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/newedit/cliutil"
)

//elvdoc:fn exit-binding
//
// Exits the current binding handler. Internally, this works by raising a
// special exception.

func exitBinding() error {
	return cliutil.ActionError(clitypes.NoAction)
}

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

func commitCode() error {
	return cliutil.ActionError(clitypes.CommitCode)
}

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func commitEOF() error {
	return cliutil.ActionError(clitypes.CommitEOF)
}
