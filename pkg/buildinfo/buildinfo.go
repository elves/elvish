// Package buildinfo contains build information.
//
// Exported string variables may be set during compilation using a linker flag
// like this:
//
//	go build -ldflags '-X src.elv.sh/pkg/buildinfo.NAME=VALUE' ./cmd/elvish
//
// This mechanism can be used by packagers to enhance Elvish's version
// information. The variables that can be set are documented below.
//
// # BuildVariant
//
// [BuildVariant], if non-empty, gets appended to the version string along with
// a "+" prefix. It should be set to a value identifying the build environment.
//
// Typically, this should be the name of the software distribution that is
// packaging Elvish, possibly plus the revision of the package. Example for
// revision 1 of a Debian package:
//
//	go build -ldflags '-X src.elv.sh/pkg/buildinfo.BuildVariant=deb1' ./cmd/elvish
//
// Supposing that [VersionBase] is "0.233.0", this causes "elvish -version" to
// print out "0.233.0+deb1".
//
// The value "official" is reserved for official binaries linked from
// https://elv.sh/get. Do not use it unless you can ensure that your build is
// bit-to-bit identical with the official binaries and you are committing to
// maintaining that property.
//
// # VCSOverride
//
// On development commits, Elvish uses the information from Git to generate a
// version string like (following the format of [Go module pseudo-versions]):
//
//	0.234.0-dev.0.20220320172241-5dc8c02a32cf
//
// where 20220320172241 is the commit time (in YYYYMMDDHHMMSS) and 5dc8c02a32cf
// is the first 12 digits of the commit hash.
//
// If this information is not available when Elvish was built - for example, if
// the build works from an archive of the commit rather than a Git checkout -
// the version string will instead look like this:
//
//	0.234.0-dev.unknown
//
// In that case, [VCSOverride] can be set to to supply the $time-$commit
// information:
//
//	go build -ldflags '-X src.elv.sh/pkg/buildinfo.VCSOverride=20220320172241-5dc8c02a32cf' ./cmd/elvish
//
// Setting this variable is only necessary when building development commits and
// the VCS information is not available.
//
// [Go module pseudo-versions]: https://go.dev/ref/mod#pseudo-versions
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

// VersionBase identifies the version of Elvish.
//
//   - On release branches, it identifies the exact version of the commit, and
//     is consistent with the Git tag. For example, at tag v0.233.0, this will
//     be "0.233.0".
//
//   - On development branches, it identifies the first version of the next
//     release branch. For example, after releases for 0.233.x has been branched
//     but before 0.234.x is branched, this will be "0.244.0". The full version
//     string will be augmented with VCS information (see [VCSOverride]).
//
// In both cases, the full version is also augmented with the [BuildVariant].
const VersionBase = "0.22.0"

// VCSOverride may be set to identify the commit of development builds when that
// information is not available during build time. It has no effect on release
// branches. See the package godoc for more details.
var VCSOverride string

// BuildVariant may be set to identify the build environment. See the package
// godoc for more details.
var BuildVariant string

// Type contains all the build information fields.
type Type struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
}

// Value contains all the build information.
var Value = Type{
	// On a release branch, change to addVariant(VersionBase, BuildVariant).
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
