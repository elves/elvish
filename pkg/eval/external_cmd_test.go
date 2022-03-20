package eval_test

import (
	"os"
	"os/exec"
	"testing"

	. "src.elv.sh/pkg/eval"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestBuiltinFnExternal(t *testing.T) {
	tmpHome := testutil.InTempHome(t)
	testutil.Setenv(t, "PATH", tmpHome+":"+os.Getenv("PATH"))

	Test(t,
		That("var e = (external true); kind-of $e").Puts("fn"),
		That("var e = (external true); put (repr $e)").Puts("<external true>"),
		That("var e = (external false); var m = [&$e=true]; put (repr $m)").Puts("[&<external false>=true]"),
		// Test calling of external commands.
		That("var e = (external true); $e").DoesNothing(),
		That("var e = (external true); $e &option").Throws(ErrExternalCmdOpts, "$e &option"),
		That("var e = (external false); $e").Throws(CmdExit(
			ExternalCmdExit{CmdName: "false", WaitStatus: exitWaitStatus(1)})),

		// TODO: Modify the ExternalCmd.Call method to wrap the Go error in a
		// predictable Elvish error so we don't have to resort to using
		// ThrowsAny in the following tests.
		//
		// The command shouldn't be found when run so we should get an
		// exception along the lines of "executable file not found in $PATH".
		That("var e = (external true); { tmp E:PATH = /; $e }").
			Throws(ErrorWithType(&exec.Error{})),
	)
}
