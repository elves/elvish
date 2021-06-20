// Elvish is a cross-platform shell, supporting Linux, BSDs and Windows. It
// features an expressive programming language, with features like namespacing
// and anonymous functions, and a fully programmable user interface with
// friendly defaults. It is suitable for both interactive use and scripting.
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
		buildinfo.Program, daemonStub{}, shell.Program{}))
}

var errNoDaemon = errors.New("daemon is not supported in this build")

type daemonStub struct{}

func (daemonStub) ShouldRun(f *prog.Flags) bool { return f.Daemon }

func (daemonStub) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	return errNoDaemon
}
