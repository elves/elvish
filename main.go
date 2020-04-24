// Elvish is a cross-platform shell, supporting Linux, BSDs and Windows. It
// features an expressive programming language, with features like namespacing
// and anonymous functions, and a fully programmable user interface with
// friendly defaults. It is suitable for both interactive use and scripting.
package main

import (
	"os"

	"github.com/elves/elvish/pkg/prog"
)

func main() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr},
		os.Args,
		prog.VersionProgram{}, prog.BuildInfoProgram{}, prog.DaemonProgram{},
		prog.WebProgram{}, prog.ShellProgram{}))
}
