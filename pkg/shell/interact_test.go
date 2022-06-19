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

func TestInteract_Eval(t *testing.T) {
	setupCleanHomePaths(t)
	testutil.InTempDir(t)
	testutil.MustWriteFile("rc.elv", "echo hello from rc.elv")
	testutil.MustWriteFile("rc-dnc.elv", "echo $a")
	testutil.MustWriteFile("rc-fail.elv", "fail bad")

	Test(t, &Program{},
		thatElvishInteract().WithStdin("echo hello\n").WritesStdout("hello\n"),
		thatElvishInteract().WithStdin("fail mock\n").WritesStderrContaining("fail mock"),

		thatElvishInteract("-rc", "rc.elv").WritesStdout("hello from rc.elv\n"),
		// rc file does not compile
		thatElvishInteract("-rc", "rc-dnc.elv").
			WritesStderrContaining("variable $a not found"),
		// rc file throws exception
		thatElvishInteract("-rc", "rc-fail.elv").WritesStderrContaining("fail bad"),
		// rc file not existing is OK
		thatElvishInteract("-rc", "rc-nonexistent.elv").DoesNothing(),
	)
}

func TestInteract_RCPath_Legacy(t *testing.T) {
	home := setupCleanHomePaths(t)
	// Legacy RC path
	testutil.MustWriteFile(
		filepath.Join(home, ".elvish", "rc.elv"), "echo hello legacy rc.elv")
	// Note: non-legacy path is tested in interact_unix_test.go

	Test(t, &Program{},
		thatElvishInteract().
			WritesStdout("hello legacy rc.elv\n").
			WritesStderrContaining(legacyRcPathWarning),
	)
}

func TestInteract_RCPath_XDG_CONFIG_HOME(t *testing.T) {
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

func thatElvishInteract(args ...string) Case {
	return ThatElvish(args...).WritesStderrContaining("")
}
