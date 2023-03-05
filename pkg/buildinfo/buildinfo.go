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
	"runtime/debug"
	"strings"
	"time"

	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/prog"
)

// VersionBase identifies the version of Elvish. On the development branches, it
// identifies the next release.
const VersionBase = "0.20.0"

// VCSOverride may be set during compilation to "time-commit" (e.g.
// "20220320172241-5dc8c02a32cf") for identifying the version of development
// builds.
//
// It is only needed if the automatic population of version information
// implemented in devVersion fails.
var VCSOverride string

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
	Version:   addVariant(devVersion(VersionBase, VCSOverride), BuildVariant),
	GoVersion: runtime.Version(),
}

func addVariant(version, variant string) string {
	if variant != "" {
		version += "+" + variant
	}
	return version
}

var readBuildInfo = debug.ReadBuildInfo

func devVersion(next, vcsOverride string) string {
	if vcsOverride != "" {
		return next + "-dev.0." + vcsOverride
	}
	fallback := next + "-dev.unknown"
	bi, ok := readBuildInfo()
	if !ok {
		return fallback
	}
	// If the main module's version is known, use it, but without the "v"
	// prefix. This is the case when Elvish is built with "go install
	// src.elv.sh/cmd/elvish@version".
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		return strings.TrimPrefix(v, "v")
	}
	// If VCS information is available (i.e. when Elvish is built from a checked
	// out repo), build the version string with it. Emulate the format of pseudo
	// version (https://go.dev/ref/mod#pseudo-versions).
	m := make(map[string]string)
	for _, s := range bi.Settings {
		if k := strings.TrimPrefix(s.Key, "vcs."); k != s.Key {
			m[k] = s.Value
		}
	}
	if m["revision"] == "" || m["time"] == "" || m["modified"] == "" {
		return fallback
	}
	t, err := time.Parse(time.RFC3339Nano, m["time"])
	if err != nil {
		return fallback
	}
	revision := m["revision"]
	if len(revision) > 12 {
		revision = revision[:12]
	}
	version := fmt.Sprintf("%s-dev.0.%s-%s", next, t.Format("20060102150405"), revision)
	if m["modified"] == "true" {
		return version + "-dirty"
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
