package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/cliutil"
)

//elvdoc:fn commit-code
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. Internally, this works by raising a special exception.

func commitCode(ev cli.KeyEvent) error {
	ev.CommitCode()
	return cliutil.ActionError(clitypes.CommitCode)
}

//elvdoc:fn commit-eof
//
// Causes the Elvish REPL to terminate. Internally, this works by raising a
// special exception.

func commitEOF() error {
	return cliutil.ActionError(clitypes.CommitEOF)
}
