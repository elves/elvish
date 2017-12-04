// +build !windows
// +build !plan9

package shell

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/edit"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/sys"
)

func makeEditor(in, out *os.File, ev *eval.Evaler, daemon *api.Client) *edit.Editor {
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGWINCH)
	return edit.NewEditor(os.Stdin, os.Stderr, sigch, ev, daemon)
}

func handleSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGHUP:
		syscall.Kill(0, syscall.SIGHUP)
		os.Exit(0)
	case syscall.SIGUSR1:
		fmt.Print(sys.DumpStack())
	}
}
