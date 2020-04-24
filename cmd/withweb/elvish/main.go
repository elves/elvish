// Command elvish is an alternative main program of Elvish that includes the web
// subprogram.
package main

import (
	"os"

	"github.com/elves/elvish/pkg/buildinfo"
	"github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/shell"
	"github.com/elves/elvish/pkg/web"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		buildinfo.Program, daemon.Program, web.Program, shell.Program))
}
