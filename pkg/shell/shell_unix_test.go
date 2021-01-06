// +build !windows,!plan9,!js

package shell

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/testutil"

	. "github.com/elves/elvish/pkg/prog/progtest"
)

func TestShell_ConnectsToDaemon(t *testing.T) {
	f := Setup()
	defer f.Cleanup()

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
	case <-time.After(testutil.ScaledMs(100)):
		t.Fatalf("timed out waiting for daemon to start")
	}

	// This test uses Script, but it also applies to Interact since the daemon
	// connection logic is common to both modes.
	Script(f.Fds(),
		[]string{"use daemon; print $daemon:pid"},
		&ScriptConfig{
			Cmd: true, SpawnDaemon: true,
			Paths: Paths{Sock: "sock", Db: "db", RunDir: "."}})
	f.TestOut(t, 1, strconv.Itoa(os.Getpid()))
	f.TestOut(t, 2, "")
}
