package daemon

import (
	"io"
	"testing"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/testutil"
)

func TestActivate_WhenServerExists(t *testing.T) {
	setup(t)
	startServer(t)
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
}

func TestActivate_FailsIfSockExistsAndIsNotSocket(t *testing.T) {
	setup(t)
	testutil.MustCreateEmpty("sock")
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err == nil {
		t.Errorf("got error nil, want non-nil")
	}
}
