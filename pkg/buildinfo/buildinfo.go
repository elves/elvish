// Package buildinfo contains build information.
//
// Build information should be set during compilation by passing
// -ldflags "-X src.elv.sh/pkg/buildinfo.Var=value" to "go build" or
// "go get".
package buildinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"src.elv.sh/pkg/prog"
)

// Version identifies the version of Elvish. On development commits, it
// identifies the next release.
const Version = "0.16.0"

// VersionSuffix is appended to Version to build the full version string. It is public so it can be
// overridden when building Elvish; see PACKAGING.md for details.
var VersionSuffix = "-dev.unknown"

// Reproducible identifies whether the build is reproducible. This can be
// overridden when building Elvish; see PACKAGING.md for details.
var Reproducible = "false"

// Program is the buildinfo subprogram.
var Program prog.Program = program{}

type Buildinfo struct {
	Version      string `json:"version"`
	Reproducible bool   `json:"reproducible"`
	GoVersion    string `json:"goversion"`
}

func (Buildinfo) IsStructMap() {}

type program struct{}

func (program) ShouldRun(f *prog.Flags) bool { return f.Version || f.BuildInfo }

func GetBuildInfo() Buildinfo {
	return Buildinfo{
		Version:      Version + VersionSuffix,
		Reproducible: Reproducible == "true",
		GoVersion:    runtime.Version(),
	}
}

func (program) Run(fds [3]*os.File, f *prog.Flags, _ []string) error {
	bi := GetBuildInfo()
	switch {
	case f.JSON:
		// Note: The only way this will be executed is if option -buildinfo is also present.
		// See the ShouldRun() method above.
		b, err := json.Marshal(bi)
		if err != nil {
			panic("unexpected buildinfo related error")
		}
		fmt.Fprintln(fds[1], string(b))
	case f.BuildInfo:
		fmt.Fprintln(fds[1], "Version:", bi.Version)
		fmt.Fprintln(fds[1], "Go version:", bi.GoVersion)
		fmt.Fprintln(fds[1], "Reproducible build:", bi.Reproducible)
	case f.Version:
		fmt.Fprintln(fds[1], bi.Version)
	default:
		panic("unexpected buildinfo related error")
	}
	return nil
}
