//go:build !windows && !plan9 && !js

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/env"

	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestInteract_NewRcFile_Default(t *testing.T) {
	home := setupCleanHomePaths(t)
	testutil.MustWriteFile(
		filepath.Join(home, ".config", "elvish", "rc.elv"), "echo hello new rc.elv")

	Test(t, &Program{},
		thatElvishInteract().WritesStdout("hello new rc.elv\n"),
	)
}

func TestInteract_NewRcFile_XDG_CONFIG_HOME(t *testing.T) {
	setupCleanHomePaths(t)
	xdgConfigHome := testutil.Setenv(t, env.XDG_CONFIG_HOME, testutil.TempDir(t))
	testutil.MustWriteFile(
		filepath.Join(xdgConfigHome, "elvish", "rc.elv"),
		"echo hello XDG_CONFIG_HOME rc.elv")

	Test(t, &Program{},
		thatElvishInteract().WritesStdout("hello XDG_CONFIG_HOME rc.elv\n"),
	)
}

func TestInteract_ConnectsToDaemon(t *testing.T) {
	testutil.InTempDir(t)

	// Run the daemon in the same process for simplicity.
	daemonDone := make(chan struct{})
	defer func() {
		select {
		case <-daemonDone:
		case <-time.After(testutil.Scaled(2 * time.Second)):
			t.Errorf("timed out waiting for daemon to quit")
		}
	}()
	readyCh := make(chan struct{})
	go func() {
		daemon.Serve("sock", "db", daemon.ServeOpts{Ready: readyCh})
		close(daemonDone)
	}()
	select {
	case <-readyCh:
		// Do nothing
	case <-time.After(testutil.Scaled(2 * time.Second)):
		t.Fatalf("timed out waiting for daemon to start")
	}

	Test(t, &Program{ActivateDaemon: daemon.Activate},
		thatElvishInteract("-sock", "sock", "-db", "db").
			WithStdin("use daemon; echo $daemon:pid\n").
			WritesStdout(fmt.Sprintln(os.Getpid())),
	)
}
