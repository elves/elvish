// Package buildinfo contains build information.
//
// Build information should be set during compilation by passing
// -ldflags "-X src.elv.sh/pkg/buildinfo.Var=value" to "go build" or
// "go get".
package buildinfo

import (
	"fmt"
	"os"
	"runtime"

	"src.elv.sh/pkg/prog"
)

// Build information.
var (
	Version      = "unknown"
	Reproducible = "false"
)

// Program is the buildinfo subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) ShouldRun(f *prog.Flags) bool { return f.Version || f.BuildInfo }

func (program) Run(fds [3]*os.File, f *prog.Flags, _ []string) error {
	if f.Version {
		fmt.Fprintln(fds[1], Version)
		return nil
	}
	if f.JSON {
		fmt.Fprintf(fds[1],
			`{"version":%s,"goversion":%s,"reproducible":%v}`+"\n",
			quoteJSON(Version), quoteJSON(runtime.Version()), Reproducible)
	} else {
		fmt.Fprintln(fds[1], "Version:", Version)
		fmt.Fprintln(fds[1], "Go version:", runtime.Version())
		fmt.Fprintln(fds[1], "Reproducible build:", Reproducible)
	}
	return nil
}
