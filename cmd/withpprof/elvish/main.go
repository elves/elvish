// Command elvish is an alternative main program of Elvish that supports writing
// pprof profiles.
package main

import (
	"os"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/lsp"
	"src.elv.sh/pkg/pprof"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		prog.Composite(
			&pprof.Program{}, &buildinfo.Program{}, &daemon.Program{}, &lsp.Program{},
			&shell.Program{ActivateDaemon: daemon.Activate})))
}
