package platform

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

func TestPlatform(t *testing.T) {
	testutil.Set(t, &osHostname, func() (string, error) {
		return "mach1.domain.tld", nil
	})

	defaultConfigHome, _ := fsutil.ConfigHome()
	defaultDataHome, _ := fsutil.DataHome()
	defaultStateHome, _ := fsutil.StateHome()

	TestWithSetup(t, setup,
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
		That(`put $platform:config-home`).Puts(defaultConfigHome),
		That(`put $platform:data-home`).Puts(defaultDataHome),
		That(`put $platform:state-home`).Puts(defaultStateHome),
	)

	testutil.Setenv(t, env.XDG_CONFIG_HOME, "/config-home")
	testutil.Setenv(t, env.XDG_DATA_HOME, "/data-home")
	testutil.Setenv(t, env.XDG_STATE_HOME, "/state-home")
	expectedConfigHome := parse.Quote(filepath.Join(string(filepath.Separator),
		"config-home", "elvish"))
	expectedDataHome := parse.Quote(filepath.Join(string(filepath.Separator),
		"data-home", "elvish"))
	expectedStateHome := parse.Quote(filepath.Join(string(filepath.Separator),
		"state-home", "elvish"))
	TestWithSetup(t, setup,
		That(fmt.Sprintf(`eq %s $platform:config-home`, expectedConfigHome)).Puts(true),
		That(fmt.Sprintf(`eq %s $platform:data-home`, expectedDataHome)).Puts(true),
		That(fmt.Sprintf(`eq %s $platform:state-home`, expectedStateHome)).Puts(true),
	)
}

func TestPlatform_HostNameError(t *testing.T) {
	errNoHostname := errors.New("hostname cannot be determined")

	testutil.Set(t, &osHostname, func() (string, error) {
		return "", errNoHostname
	})
	TestWithSetup(t, setup,
		That(`platform:hostname`).Throws(errNoHostname),
	)
}

func setup(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("platform", Ns))
}
