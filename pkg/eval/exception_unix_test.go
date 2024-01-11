//go:build unix

package eval_test

import (
	"fmt"
	"runtime"
	"syscall"
	"testing"

	. "src.elv.sh/pkg/eval"

	"src.elv.sh/pkg/tt"
)

func TestExternalCmdExit_Error(t *testing.T) {
	tt.Test(t, error.Error,
		Args(ExternalCmdExit{0x0, "ls", 1}).Rets("ls exited with 0"),
		Args(ExternalCmdExit{0x100, "ls", 1}).Rets("ls exited with 1"),
		// Note: all Unix'es have SIGINT = 2, but syscall package has different
		// string in gccgo("Interrupt") and gc("interrupt").
		Args(ExternalCmdExit{0x2, "ls", 1}).Rets("ls killed by signal "+syscall.SIGINT.String()),
		// 0x80 + signal for core dumped
		Args(ExternalCmdExit{0x82, "ls", 1}).Rets("ls killed by signal "+syscall.SIGINT.String()+" (core dumped)"),
		// 0x7f + signal<<8 for stopped
		Args(ExternalCmdExit{0x27f, "ls", 1}).Rets("ls stopped by signal "+syscall.SIGINT.String()+" (pid=1)"),
	)
	if runtime.GOOS == "linux" {
		tt.Test(t, error.Error,
			// 0x057f + cause<<16 for trapped. SIGTRAP is 5 on all Unix'es but have
			// different string representations on different OSes.
			Args(ExternalCmdExit{0x1057f, "ls", 1}).Rets(fmt.Sprintf(
				"ls stopped by signal %s (pid=1) (trapped 1)", syscall.SIGTRAP)),
			// 0xff is the only exit code that is not exited, signaled or stopped.
			Args(ExternalCmdExit{0xff, "ls", 1}).Rets("ls has unknown WaitStatus 255"),
		)
	}
}
