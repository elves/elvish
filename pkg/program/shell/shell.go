// Package shell is the entry point for the terminal interface of Elvish.
package shell

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/sys"
	"github.com/elves/elvish/pkg/util"
)

var logger = util.GetLogger("[shell] ")

func setupShell(fds [3]*os.File, p *Paths, spawn bool) (*eval.Evaler, func()) {
	restoreTTY := term.SetupGlobal()
	ev := InitRuntime(fds[2], p, spawn)

	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			logger.Println("signal", sig)
			handleSignal(sig, fds[2])
		}
	}()

	return ev, func() {
		signal.Stop(sigs)
		CleanupRuntime(fds[2], ev)
		restoreTTY()
	}
}

// Global panic handler.
func rescue() {
	r := recover()
	if r != nil {
		println()
		print(sys.DumpStack())
		println()
		println(r)
		println("\nExecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}
