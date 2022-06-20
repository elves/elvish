package shell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/testutil"
)

func TestInteract_Eval(t *testing.T) {
	setupCleanHomePaths(t)
	testutil.InTempDir(t)
	must.WriteFile("rc.elv", "echo hello from rc.elv")
	must.WriteFile("rc-dnc.elv", "echo $a")
	must.WriteFile("rc-fail.elv", "fail bad")

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
	must.WriteFile(
		filepath.Join(home, ".elvish", "rc.elv"), "echo hello legacy rc.elv")

	Test(t, &Program{},
		thatElvishInteract().
			WritesStdout("hello legacy rc.elv\n").
			WritesStderrContaining(legacyRcPathWarning),
	)
}

func TestInteract_RCPath_XDG_CONFIG_HOME(t *testing.T) {
	setupCleanHomePaths(t)
	xdgConfigHome := testutil.Setenv(t, env.XDG_CONFIG_HOME, testutil.TempDir(t))
	must.WriteFile(
		filepath.Join(xdgConfigHome, "elvish", "rc.elv"),
		"echo hello XDG_CONFIG_HOME rc.elv")

	Test(t, &Program{},
		thatElvishInteract().WritesStdout("hello XDG_CONFIG_HOME rc.elv\n"),
	)
}

func TestInteract_ConnectsToDaemon(t *testing.T) {
	sockPath := startDaemon(t)

	Test(t, &Program{ActivateDaemon: fakeActivate(sockPath)},
		thatElvishInteract().
			WithStdin("use daemon; echo $daemon:pid\n").
			WritesStdout(fmt.Sprintln(os.Getpid())),
	)
}

func TestInteract_DoesNotStoreEmptyCommandInHistory(t *testing.T) {
	sockPath := startDaemon(t)
	Test(t, &Program{ActivateDaemon: fakeActivate(sockPath)},
		thatElvishInteract().
			WithStdin("\n"+"use store; print (store:next-cmd-seq)\n").
			WritesStdout("1"),
	)
}

func TestInteract_ErrorInActivateDaemon(t *testing.T) {
	activate := func(io.Writer, *daemondefs.SpawnConfig) (daemondefs.Client, error) {
		return nil, errors.New("fake error")
	}
	Test(t, &Program{ActivateDaemon: activate},
		thatElvishInteract().
			WritesStderrContaining("Cannot connect to daemon: fake error"),
	)
}

func TestInteract_DBPath_Legacy(t *testing.T) {
	sockPath := startDaemon(t)
	home := setupCleanHomePaths(t)
	legacyDBPath := filepath.Join(home, ".elvish", "db")
	must.WriteFile(legacyDBPath, "")

	Test(t, &Program{ActivateDaemon: fakeActivate(sockPath)},
		thatElvishInteract().
			WritesStderrContaining("db requested: "+legacyDBPath),
	)
}

func TestInteract_DBPath_XDG_STATE_HOME(t *testing.T) {
	sockPath := startDaemon(t)
	setupCleanHomePaths(t)
	xdgStateHome := testutil.Setenv(t, env.XDG_STATE_HOME, t.TempDir())

	Test(t, &Program{ActivateDaemon: fakeActivate(sockPath)},
		thatElvishInteract().
			WritesStderrContaining("db requested: "+
				filepath.Join(xdgStateHome, "elvish", "db.bolt")),
	)
}

func thatElvishInteract(args ...string) Case {
	return ThatElvish(args...).WritesStderrContaining("")
}

// Starts a daemon, and returns the socket path to connect it with.
func startDaemon(t *testing.T) string {
	t.Helper()
	// Run the daemon in the same process for simplicity.
	dir := testutil.TempDir(t)
	sockPath := filepath.Join(dir, "sock")
	sigCh := make(chan os.Signal)
	readyCh := make(chan struct{})
	daemonDone := make(chan struct{})
	go func() {
		daemon.Serve(sockPath, filepath.Join(dir, "db.bolt"),
			daemon.ServeOpts{Ready: readyCh, Signals: sigCh})
		close(daemonDone)
	}()
	t.Cleanup(func() {
		t.Helper()
		close(sigCh)
		select {
		case <-daemonDone:
		case <-time.After(testutil.Scaled(2 * time.Second)):
			t.Errorf("timed out waiting for daemon to quit")
		}
	})
	select {
	case <-readyCh:
		// Do nothing
	case <-time.After(testutil.Scaled(2 * time.Second)):
		t.Fatalf("timed out waiting for daemon to start")
	}
	return sockPath
}

func fakeActivate(sockPath string) daemondefs.ActivateFunc {
	return func(stderr io.Writer, cfg *daemondefs.SpawnConfig) (daemondefs.Client, error) {
		fmt.Fprintln(stderr, "db requested:", cfg.DbPath)
		// Always connect to the in-process daemon just started.
		return daemon.NewClient(sockPath), nil
	}
}
