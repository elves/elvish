// +build !windows,!plan9,!js

package eval

import (
	"fmt"
	"runtime"
	"syscall"
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestExternalCmdExit_Error(t *testing.T) {
	tt.Test(t, tt.Fn("Error", error.Error), tt.Table{
		tt.Args(ExternalCmdExit{0x0, "ls", 1}).Rets("ls exited with 0"),
		tt.Args(ExternalCmdExit{0x100, "ls", 1}).Rets("ls exited with 1"),
		// Note: all Unix'es have SIGINT = 2, but syscall package has different
		// string in gccgo("Interrupt") and gc("interrupt").
		tt.Args(ExternalCmdExit{0x2, "ls", 1}).Rets("ls killed by signal " + syscall.SIGINT.String()),
		// 0x80 + signal for core dumped
		tt.Args(ExternalCmdExit{0x82, "ls", 1}).Rets("ls killed by signal " + syscall.SIGINT.String() + " (core dumped)"),
		// 0x7f + signal<<8 for stopped
		tt.Args(ExternalCmdExit{0x27f, "ls", 1}).Rets("ls stopped by signal " + syscall.SIGINT.String() + " (pid=1)"),
	})
	if runtime.GOOS == "linux" {
		tt.Test(t, tt.Fn("Error", error.Error), tt.Table{
			// 0x057f + cause<<16 for trapped. SIGTRAP is 5 on all Unix'es but have
			// different string representations on different OSes.
			tt.Args(ExternalCmdExit{0x1057f, "ls", 1}).Rets(fmt.Sprintf(
				"ls stopped by signal %s (pid=1) (trapped 1)", syscall.SIGTRAP)),
			// 0xff is the only exit code that is not exited, signaled or stopped.
			tt.Args(ExternalCmdExit{0xff, "ls", 1}).Rets("ls has unknown WaitStatus 255"),
		})
	}
}
