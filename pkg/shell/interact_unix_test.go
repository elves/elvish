//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/env"

	. "src.elv.sh/pkg/prog/progtest"
	. "src.elv.sh/pkg/testutil"
)

func TestInteract_NewRcFile_Default(t *testing.T) {
	home := setupHomePaths(t)
	MustWriteFile(
		filepath.Join(home, ".config", "elvish", "rc.elv"), "echo hello new rc.elv")

	Test(t, Program{},
		thatElvishInteract().WritesStdout("hello new rc.elv\n"),
	)
}

func TestInteract_NewRcFile_XDG_CONFIG_HOME(t *testing.T) {
	setupHomePaths(t)
	xdgConfigHome := Setenv(t, env.XDG_CONFIG_HOME, TempDir(t))
	MustWriteFile(
		filepath.Join(xdgConfigHome, "elvish", "rc.elv"),
		"echo hello XDG_CONFIG_HOME rc.elv")

	Test(t, Program{},
		thatElvishInteract().WritesStdout("hello XDG_CONFIG_HOME rc.elv\n"),
	)
}

func TestInteract_ConnectsToDaemon(t *testing.T) {
	InTempDir(t)

	// Run the daemon in the same process for simplicity.
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()
	go func() {
		daemon.Serve("sock", "db")
		wg.Done()
	}()
	// Block until the socket file exists, so that Elvish will not try to spawn
	// again.
	hasSock := make(chan struct{})
	go func() {
		defer close(hasSock)
		for {
			_, err := os.Stat("sock")
			if err == nil {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	select {
	case <-hasSock:
		// Do nothing
	case <-time.After(ScaledMs(100)):
		t.Fatalf("timed out waiting for daemon to start")
	}

	Test(t, Program{client.Activate},
		thatElvishInteract("-sock", "sock", "-db", "db").
			WithStdin("use daemon; echo $daemon:pid\n").
			WritesStdout(fmt.Sprintln(os.Getpid())),
	)
}
