package platform

import (
	"errors"
	"runtime"
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

func TestPlatform(t *testing.T) {
	testutil.Set(t, &osHostname, func() (string, error) {
		return "mach1.domain.tld", nil
	})

	TestWithEvalerSetup(t, setup,
		That(`put $platform:arch`).Puts(runtime.GOARCH),
		That(`put $platform:os`).Puts(runtime.GOOS),
		That(`put $platform:is-windows`).Puts(runtime.GOOS == "windows"),
		That(`put $platform:is-unix`).Puts(
			// Convert to bool type explicitly, to workaround gccgo bug.
			// https://github.com/golang/go/issues/40152
			// TODO(zhsj): remove workaround after gcc 11 is the default in CI.
			bool(runtime.GOOS != "windows" && runtime.GOOS != "plan9" && runtime.GOOS != "js")),
		That(`platform:hostname`).Puts("mach1.domain.tld"),
		That(`platform:hostname &strip-domain`).Puts("mach1"),
	)
}

func TestPlatform_HostNameError(t *testing.T) {
	errNoHostname := errors.New("hostname cannot be determined")

	testutil.Set(t, &osHostname, func() (string, error) {
		return "", errNoHostname
	})
	TestWithEvalerSetup(t, setup,
		That(`platform:hostname`).Throws(errNoHostname),
	)
}

func setup(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("platform", Ns))
}
