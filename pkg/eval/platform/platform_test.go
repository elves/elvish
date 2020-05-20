package platform

import (
	"errors"
	"runtime"
	"testing"

	"github.com/elves/elvish/pkg/eval"
)

const (
	testHostname = "mach1.domain.tld"
	testMachname = "mach1"
)

var hostnameFail = true

func hostnameMock() (string, error) {
	if hostnameFail {
		hostnameFail = false
		return "", errors.New("hostname cannot be determined")
	}
	return testHostname, nil
}

var That = eval.That

func TestPlatform(t *testing.T) {
	savedOsHostname := osHostname
	osHostname = hostnameMock
	defer func() { osHostname = savedOsHostname }()
	setup := func(ev *eval.Evaler) { ev.Builtin.AddNs("platform", Ns) }
	eval.TestWithSetup(t, setup,
		That(`put $platform:arch`).Puts(runtime.GOARCH),
		That(`put $platform:os`).Puts(runtime.GOOS),
		That(`put $platform:is-windows`).Puts(runtime.GOOS == "windows"),
		That(`put $platform:is-unix`).Puts(
			runtime.GOOS != "windows" && runtime.GOOS != "plan9" && runtime.GOOS != "js"),
		// The first time we invoke the mock it acts as if we can't determine
		// the hostname. Make sure that is turned into the expected exception.
		That(`platform:hostname`).ThrowsCause(errCannotDetermineHostname),
		That(`platform:hostname`).Puts(testHostname),
		That(`platform:hostname &strip-domain`).Puts(testMachname),
	)
}
