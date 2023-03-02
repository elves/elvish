//go:build unix

package daemon

import (
	"io"
	"os"
	"os/user"
	"testing"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/must"
)

func TestActivate_InterruptsOutdatedServerAndSpawnsNewServer(t *testing.T) {
	activated := 0
	setupForActivate(t, func(name string, argv []string, attr *os.ProcAttr) error {
		startServer(t, argv)
		activated++
		return nil
	})
	version := api.Version - 1
	oldServer := startServerOpts(t, cli("sock", "db"), ServeOpts{Version: &version})

	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
	if activated != 1 {
		t.Errorf("got activated %v times, want 1", activated)
	}
	oldServer.WaitQuit()
}

func TestActivate_FailsIfUnableToRemoveHangingSocket(t *testing.T) {
	if u, err := user.Current(); err != nil || u.Uid == "0" {
		t.Skip("current user is root or unknown")
	}
	activated := 0
	setupForActivate(t, func(name string, argv []string, attr *os.ProcAttr) error {
		activated++
		return nil
	})
	must.MkdirAll("d")
	makeHangingUnixSocket(t, "d/sock")
	// Remove write permission so that removing d/sock will fail
	os.Chmod("d", 0600)
	defer os.Chmod("d", 0700)

	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "d/sock", RunDir: "."})
	if err == nil {
		t.Errorf("got error nil, want non-nil")
	}
	if activated != 0 {
		t.Errorf("got activated %v times, want 0", activated)
	}
}
