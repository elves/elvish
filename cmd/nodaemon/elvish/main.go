// Command elvish is an alternative main program of Elvish that does not include
// the daemon subprogram.
package main

import (
	"errors"
	"os"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		prog.Composite(buildinfo.Program, daemonStub{}, shell.Program{})))
}

var errNoDaemon = errors.New("daemon is not supported in this build")

type daemonStub struct{}

func (daemonStub) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	if f.Daemon {
		return errNoDaemon
	}
	return prog.ErrNotSuitable
}
