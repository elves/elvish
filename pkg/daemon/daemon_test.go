package daemon

import (
	"syscall"
	"testing"
	"time"

	"github.com/elves/elvish/pkg/testutil"
	"github.com/elves/elvish/pkg/util"
)

func TestDaemon(t *testing.T) {
	// Set up filesystem.
	_, cleanup := util.InTestDir()
	defer cleanup()

	// Set up server.
	serverDone := make(chan struct{})
	go func() {
		Serve("sock", "db")
		close(serverDone)
	}()
	defer func() { <-serverDone }()

	// Set up client.
	client := NewClient("sock")
	defer client.Close()
	for i := 0; i < 100; i++ {
		client.ResetConn()
		_, err := client.Version()
		if err == nil {
			break
		} else if i == 99 {
			t.Fatal("Failed to connect after 1s")
		}
		time.Sleep(testutil.ScaledMs(10))
	}

	// Server state requests.

	gotVersion, err := client.Version()
	if gotVersion != Version || err != nil {
		t.Errorf(".Version() -> (%v, %v), want (%v, nil)", gotVersion, err, Version)
	}

	gotPid, err := client.Pid()
	wantPid := syscall.Getpid()
	if gotPid != wantPid || err != nil {
		t.Errorf(".Pid() -> (%v, %v), want (%v, nil)", gotPid, err, wantPid)
	}

	// Store requests.

	_, err = client.AddCmd("test cmd")
	if err != nil {
		t.Errorf("client.AddCmd -> error %v", err)
	}
}
