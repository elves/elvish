// Package platform exposes variables and functions that deal with the
// specific platform being run on, such as the OS name and CPU architecture.
package platform

import (
	"os"
	"runtime"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
)

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

var Ns = eval.BuildNsNamed("platform").
	AddVars(map[string]vars.Var{
		"arch":       vars.NewReadOnly(runtime.GOARCH),
		"os":         vars.NewReadOnly(runtime.GOOS),
		"is-unix":    vars.NewReadOnly(isUnix),
		"is-windows": vars.NewReadOnly(isWindows),
	}).
	AddGoFns(map[string]any{
		"hostname": hostname,
	}).Ns()
