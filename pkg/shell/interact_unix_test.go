//go:build !windows && !plan9 && !js

package shell

import (
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"

	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestInteract_RCPath_Default(t *testing.T) {
	home := setupCleanHomePaths(t)
	testutil.Unsetenv(t, env.XDG_CONFIG_HOME)
	testutil.MustWriteFile(
		filepath.Join(home, ".config", "elvish", "rc.elv"), "echo hello new rc.elv")

	Test(t, &Program{},
		thatElvishInteract().WritesStdout("hello new rc.elv\n"),
	)
}
