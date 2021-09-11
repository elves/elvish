// Command elvish is an alternative main program of Elvish that includes the web
// subprogram.
package main

import (
	"os"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
	"src.elv.sh/pkg/web"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		prog.Composite(
			buildinfo.Program, daemon.Program, web.Program,
			shell.Program{ActivateDaemon: client.Activate})))
}
