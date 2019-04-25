package editutil

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/newedit/clitypes"
	"github.com/elves/elvish/tt"
)

var Args = tt.Args

func TestActionError_Error(t *testing.T) {
	tt.Test(t, tt.Fn("ActionError.Error", ActionError.Error), tt.Table{
		Args(ActionError(clitypes.CommitCode)).Rets("commit-code"),
		Args(ActionError(-1)).Rets("!(BAD ACTION: -1)"),
	})
}

func TestActionError_Repr(t *testing.T) {
	tt.Test(t, tt.Fn("Repr", vals.Repr), tt.Table{
		Args(ActionError(clitypes.CommitCode), 0).Rets("?(edit:commit-code)"),
	})
}

func TestActionError_PPrint(t *testing.T) {
	tt.Test(t, tt.Fn("ActionError.PPrint", ActionError.PPrint), tt.Table{
		Args(ActionError(clitypes.CommitCode), "").
			Rets("\033[33;1mcommit-code\033[m"),
	})
}
