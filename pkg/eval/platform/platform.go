// Package platform exposes variables and functions that deal with the
// specific platform being run on, such as the OS name and CPU architecture.
package platform

import (
	"errors"
	"os"
	"runtime"
	"strings"

	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vars"
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

//elvdoc:fn hostname &strip-domain=false
//
// The host name of the system. If option `&strip-domain` is used any domain
// components are stripped from the host name. Domain components include the
// first dot (period) and anything that follows. For example,
// `machname.subdomain.tld` => `machname`.
//
// If an error occurs while fetching the host name an exception is thrown.
// This is implemented using Go's
// [`os.Hostname`](https://golang.org/pkg/os/#Hostname) function.

var osHostname = os.Hostname // to allow mocking in unit tests
var errCannotDetermineHostname = errors.New("cannot determine the hostname")

type hostnameOpt struct{ StripDomain bool }

func (o *hostnameOpt) SetDefaultOptions() {}

func hostname(opts hostnameOpt) (string, error) {
	hostname, err := osHostname()
	if err != nil {
		return "", errCannotDetermineHostname
	}
	if !opts.StripDomain {
		return hostname, nil
	}
	parts := strings.SplitN(hostname, ".", 2)
	return parts[0], nil
}

var Ns = eval.Ns{
	"arch":       vars.NewReadOnly(runtime.GOARCH),
	"os":         vars.NewReadOnly(runtime.GOOS),
	"is-unix":    vars.NewReadOnly(isUnix),
	"is-windows": vars.NewReadOnly(isWindows),
}.AddGoFns("platform:", map[string]interface{}{
	"hostname": hostname,
})
