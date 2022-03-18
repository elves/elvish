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
const Version = "0.18.0"

// VersionSuffix is appended to Version to build the full version string. It is public so it can be
// overridden when building Elvish; see PACKAGING.md for details.
var VersionSuffix = ""

// Reproducible identifies whether the build is reproducible. This can be
// overridden when building Elvish; see PACKAGING.md for details.
var Reproducible = "false"

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

// Program is the buildinfo subprogram.
type Program struct {
	version, buildinfo bool
	json               *bool
}

func (p *Program) RegisterFlags(fs *prog.FlagSet) {
	fs.BoolVar(&p.version, "version", false,
		"Output the Elvish version and quit")
	fs.BoolVar(&p.buildinfo, "buildinfo", false,
		"Output information about the Elvish build and quit")
	p.json = fs.JSON()
}

func (p *Program) Run(fds [3]*os.File, _ []string) error {
	switch {
	case p.buildinfo:
		if *p.json {
			fmt.Fprintln(fds[1], mustToJSON(Value))
		} else {
			fmt.Fprintln(fds[1], "Version:", Value.Version)
			fmt.Fprintln(fds[1], "Go version:", Value.GoVersion)
			fmt.Fprintln(fds[1], "Reproducible build:", Value.Reproducible)
		}
	case p.version:
		if *p.json {
			fmt.Fprintln(fds[1], mustToJSON(Value.Version))
		} else {
			fmt.Fprintln(fds[1], Value.Version)
		}
	default:
		return prog.ErrNextProgram
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
