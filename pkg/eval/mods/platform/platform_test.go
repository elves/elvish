package platform

import (
	"errors"
	"runtime"
	"testing"

	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
)

const (
	testHostname = "mach1.domain.tld"
	testMachname = "mach1"
)

var (
	hostnameFail  = true
	errNoHostname = errors.New("hostname cannot be determined")
)

func hostnameMock() (string, error) {
	if hostnameFail {
		hostnameFail = false
		return "", errNoHostname
	}
	return testHostname, nil
}

func TestPlatform(t *testing.T) {
	savedOsHostname := osHostname
	osHostname = hostnameMock
	hostnameFail = true
	defer func() { osHostname = savedOsHostname }()
	setup := func(ev *eval.Evaler) {
		ev.AddGlobal(eval.NsBuilder{}.AddNs("platform", Ns).Ns())
	}
	TestWithSetup(t, setup,
		That(`put $platform:arch`).Puts(runtime.GOARCH),
		That(`put $platform:os`).Puts(runtime.GOOS),
		That(`put $platform:is-windows`).Puts(runtime.GOOS == "windows"),
		That(`put $platform:is-unix`).Puts(
			// Convert to bool type explicitly, to workaround gccgo bug.
			// https://github.com/golang/go/issues/40152
			// TODO(zhsj): remove workaround after gcc 11 is the default in CI.
			bool(runtime.GOOS != "windows" && runtime.GOOS != "plan9" && runtime.GOOS != "js")),
		// The first time we invoke the mock it acts as if we can't determine
		// the hostname. Make sure that is turned into the expected exception.
		That(`platform:hostname`).Throws(errNoHostname),

		That(`platform:hostname`).Puts(testHostname),
		That(`platform:hostname &strip-domain`).Puts(testMachname),
	)
}
