// Package platform exposes variables and functions that deal with the
// specific platform being run on, such as the OS name and CPU architecture.
package platform

import (
	"os"
	"runtime"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
)

//elvdoc:var arch
//
// The architecture of the platform; e.g. amd64, arm, ppc.
// This corresponds to Go's
// [`GOARCH`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
// This is read-only.

//elvdoc:var os
//
// The name of the operating system; e.g. darwin (macOS), linux, etc.
// This corresponds to Go's
// [`GOOS`](https://pkg.go.dev/runtime?tab=doc#pkg-constants) constant.
// This is read-only.

//elvdoc:var is-unix
//
// Whether or not the platform is UNIX-like. This includes Linux, macOS
// (Darwin), FreeBSD, NetBSD, and OpenBSD. This can be used to decide, for
// example, if the `unix` module is usable.
// This is read-only.

//elvdoc:var is-windows
//
// Whether or not the platform is Microsoft Windows.
// This is read-only.

//elvdoc:var config-home
//
// The directory containing the *rc.elv* file as defined [here](./command.html#rc-file).
// This is read-only.

//elvdoc:var data-home
//
// The directory containing third-party Elvish modules managed by [`epm`](./epm.html) as defined
// [here](./command.html#module-search-directories).
// This is read-only.

//elvdoc:var state-home
//
// The directory containing the Elvish database file as defined
// [here](./command.html#database-file). This is a good place for modules to create files that are
// usually considered private to a particular machine running Elvish. That is, files that normally
// aren't synced to other machines.
// This is read-only.

//elvdoc:fn hostname
//
// ```elvish
// platform:hostname &strip-domain=$false
// ```
//
// Outputs the hostname of the system. If the option `&strip-domain` is `$true`,
// strips the part after the first dot.
//
// This function throws an exception if it cannot determine the hostname. It is
// implemented using Go's [`os.Hostname`](https://golang.org/pkg/os/#Hostname).
//
// Examples:
//
// ```elvish-transcript
// ~> platform:hostname
// ▶ lothlorien.elv.sh
// ~> platform:hostname &strip-domain=$true
// ▶ lothlorien
// ```

var osHostname = os.Hostname // to allow mocking in unit tests

type hostnameOpt struct{ StripDomain bool }

func (o *hostnameOpt) SetDefaultOptions() {}

func hostname(opts hostnameOpt) (string, error) {
	hostname, err := osHostname()
	if err != nil {
		return "", err
	}
	if !opts.StripDomain {
		return hostname, nil
	}
	parts := strings.SplitN(hostname, ".", 2)
	return parts[0], nil
}

func getConfigHome() any {
	cfgHome, _ := fsutil.ConfigHome()
	return cfgHome
}

func getDataHome() any {
	dataHome, _ := fsutil.DataHome()
	return dataHome
}

func getStateHome() any {
	stateHome, _ := fsutil.StateHome()
	return stateHome
}

var Ns = eval.BuildNsNamed("platform").
	AddVars(map[string]vars.Var{
		"arch":        vars.NewReadOnly(runtime.GOARCH),
		"os":          vars.NewReadOnly(runtime.GOOS),
		"is-unix":     vars.NewReadOnly(isUnix),
		"is-windows":  vars.NewReadOnly(isWindows),
		"config-home": vars.FromGet(getConfigHome),
		"data-home":   vars.FromGet(getDataHome),
		"state-home":  vars.FromGet(getStateHome),
	}).
	AddGoFns(map[string]any{
		"hostname": hostname,
	}).Ns()
