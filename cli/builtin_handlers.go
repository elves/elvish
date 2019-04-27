package cli

import (
	"github.com/elves/elvish/cli/clitypes"
	"github.com/elves/elvish/cli/cliutil"
	"github.com/elves/elvish/edit/tty"
)

// CommitEOF is an EventHandler that calls CommitEOF.
var CommitEOF = KeyEvent.CommitEOF

// CommitCode is an EventHandler that calls CommitCode.
var CommitCode = KeyEvent.CommitCode

// DefaultInsert is an EventHandler that is suitable as the default EventHandler
// of insert mode.
func DefaultInsert(ev KeyEvent) {
	action := cliutil.BasicHandler(tty.KeyEvent(ev.Key()), ev.State())
	switch action {
	case clitypes.CommitCode:
		ev.CommitCode()
	case clitypes.CommitEOF:
		ev.CommitEOF()
	}
}
