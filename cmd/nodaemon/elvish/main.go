// Command elvish is an alternative main program of Elvish that does not include
// the daemon subprogram.
package main

import (
	"os"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/lsp"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		prog.Composite(&buildinfo.Program{}, &lsp.Program{}, &shell.Program{})))
}
