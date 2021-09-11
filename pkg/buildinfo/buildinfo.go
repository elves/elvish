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
const Version = "0.17.0"

// VersionSuffix is appended to Version to build the full version string. It is public so it can be
// overridden when building Elvish; see PACKAGING.md for details.
var VersionSuffix = "-dev.unknown"

// Reproducible identifies whether the build is reproducible. This can be
// overridden when building Elvish; see PACKAGING.md for details.
var Reproducible = "false"

// Program is the buildinfo subprogram.
var Program prog.Program = program{}

// Type contains all the build information fields.
type Type struct {
	Version      string `json:"version"`
	Reproducible bool   `json:"reproducible"`
	GoVersion    string `json:"goversion"`
}

func (Type) IsStructMap() {}

// Value contains all the build information.
var Value = Type{
	Version:      Version + VersionSuffix,
	Reproducible: Reproducible == "true",
	GoVersion:    runtime.Version(),
}

type program struct{}

func (program) Run(fds [3]*os.File, f *prog.Flags, _ []string) error {
	switch {
	case f.BuildInfo:
		if f.JSON {
			fmt.Fprintln(fds[1], mustToJSON(Value))
		} else {
			fmt.Fprintln(fds[1], "Version:", Value.Version)
			fmt.Fprintln(fds[1], "Go version:", Value.GoVersion)
			fmt.Fprintln(fds[1], "Reproducible build:", Value.Reproducible)
		}
	case f.Version:
		if f.JSON {
			fmt.Fprintln(fds[1], mustToJSON(Value.Version))
		} else {
			fmt.Fprintln(fds[1], Value.Version)
		}
	default:
		return prog.ErrNotSuitable
	}
	return nil
}

func mustToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
