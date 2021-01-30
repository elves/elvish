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

// Version identifies the version of Elvish. On development commits, it
// identifies the next release.
const Version = "v0.16.0"

// VersionSuffix is appended to Version in the output of "elvish -version" and
// "elvish -buildinfo" to build the full version string. This can be overriden
// when building Elvish; see PACKAGING.md for details.
var VersionSuffix = "-dev.unknown"

// Reproducible identifies whether the build is reproducible. This can be
// overriden when building Elvish; see PACKAGING.md for details.
var Reproducible = "false"

// Program is the buildinfo subprogram.
var Program prog.Program = program{}

type program struct{}

func (program) ShouldRun(f *prog.Flags) bool { return f.Version || f.BuildInfo }

func (program) Run(fds [3]*os.File, f *prog.Flags, _ []string) error {
	fullVersion := Version + VersionSuffix
	if f.Version {
		fmt.Fprintln(fds[1], fullVersion)
		return nil
	}
	if f.JSON {
		fmt.Fprintf(fds[1],
			`{"version":%s,"goversion":%s,"reproducible":%v}`+"\n",
			quoteJSON(fullVersion), quoteJSON(runtime.Version()), Reproducible)
	} else {
		fmt.Fprintln(fds[1], "Version:", fullVersion)
		fmt.Fprintln(fds[1], "Go version:", runtime.Version())
		fmt.Fprintln(fds[1], "Reproducible build:", Reproducible)
	}
	return nil
}
