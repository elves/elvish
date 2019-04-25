package cliutil

import (
	"fmt"

	"github.com/elves/elvish/newedit/clitypes"
)

// ActionError is like HandlerAction with an Error method. It is useful as a
// control flow exception to exit early from a handler.
type ActionError clitypes.HandlerAction

var actionNames = [...]string{
	"no-action", "commit-code",
}

func (ae ActionError) Error() string {
	if ae < 0 || int(ae) >= len(actionNames) {
		return fmt.Sprintf("!(BAD ACTION: %d)", ae)
	}
	return actionNames[ae]
}

// Repr returns an Elvish representation of ae.
func (ae ActionError) Repr(int) string {
	return "?(edit:" + ae.Error() + ")"
}

// PPrint pretty-prints ae to the terminal.
func (ae ActionError) PPrint(string) string {
	return "\033[33;1m" + ae.Error() + "\033[m"
}
