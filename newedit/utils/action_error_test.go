package utils

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestActionError_Error(t *testing.T) {
	tt.Test(t, tt.Fn("ActionError.Error", ActionError.Error), tt.Table{
		Args(ActionError(types.CommitCode)).Rets("commit-code"),
		Args(ActionError(-1)).Rets("!(BAD ACTION: -1)"),
	})
}

func TestActionError_Repr(t *testing.T) {
	tt.Test(t, tt.Fn("Repr", vals.Repr), tt.Table{
		Args(ActionError(types.CommitCode), 0).Rets("?(edit:commit-code)"),
	})
}

func TestActionError_Pprint(t *testing.T) {
	tt.Test(t, tt.Fn("ActionError.Pprint", ActionError.Pprint), tt.Table{
		Args(ActionError(types.CommitCode), "").
			Rets("\033[33;1mcommit-code\033[m"),
	})
}
