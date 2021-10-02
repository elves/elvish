//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package daemon

import (
	"io"
	"os"
	"os/user"
	"testing"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/testutil"
)

func TestActivate_FailsIfUnableToRemoveHangingSocket(t *testing.T) {
	if u, err := user.Current(); err != nil || u.Uid == "0" {
		t.Skip("current user is root or unknown")
	}
	activated := 0
	setupForActivate(t, func(name string, argv []string, attr *os.ProcAttr) error {
		activated++
		return nil
	})
	testutil.MustMkdirAll("d")
	makeHangingUNIXSocket(t, "d/sock")
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
