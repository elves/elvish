//go:build unix

package shell

import (
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/must"

	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestInteract_RCPath_Default(t *testing.T) {
	home := setupCleanHomePaths(t)
	testutil.Unsetenv(t, env.XDG_CONFIG_HOME)
	must.WriteFile(
		filepath.Join(home, ".config", "elvish", "rc.elv"), "echo hello new rc.elv")

	Test(t, &Program{},
		thatElvishInteract().WritesStdout("hello new rc.elv\n"),
	)
}

func TestInteract_DBPath_Default(t *testing.T) {
	sockPath := startDaemon(t)
	home := setupCleanHomePaths(t)

	Test(t, &Program{ActivateDaemon: fakeActivate(sockPath)},
		thatElvishInteract().
			WritesStderrContaining("db requested: "+
				filepath.Join(home, ".local", "state", "elvish", "db.bolt")),
	)
}
