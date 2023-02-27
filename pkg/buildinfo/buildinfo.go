// Package buildinfo contains build information.
//
// Some of the exported fields may be set during compilation by passing -ldflags
// "-X src.elv.sh/pkg/buildinfo.Var=value" to "go build".
package buildinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/prog"
)

// VersionBase identifies the version of Elvish. On the development branches, it
// identifies the next release.
const VersionBase = "0.19.0"

// BuildVariant may be set during compilation to identify a particular
// build variant, such as a build by a specific distribution, with modified
// dependencies, or with a non-standard toolchain.
//
// If non-empty, it is appended to the version string, along with a "+" prefix.
var BuildVariant string

// Type contains all the build information fields.
type Type struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
}

func (Type) IsStructMap() {}

// Value contains all the build information.
var Value = Type{
	// On a release branch, change to addVariant(VersionBase, BuildVariant) and
	// remove unneeded code.
	Version:   addVariant(VersionBase, BuildVariant),
	GoVersion: runtime.Version(),
}

func addVariant(version, variant string) string {
	if variant != "" {
		version += "+" + variant
	}
	return version
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
		}
	case p.version:
		if *p.json {
			fmt.Fprintln(fds[1], mustToJSON(Value.Version))
		} else {
			fmt.Fprintln(fds[1], Value.Version)
		}
	default:
		return prog.NextProgram()
	}
	return nil
}

func mustToJSON(v any) string {
	return string(must.OK1(json.Marshal(v)))
}
