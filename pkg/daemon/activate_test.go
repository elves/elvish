package daemon

import (
	"io"
	"runtime"
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

func TestActivate_FailsIfCannotStatSock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	setup(t)
	testutil.MustCreateEmpty("not-dir")
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "not-dir/sock", RunDir: "."})
	if err == nil {
		t.Errorf("got error nil, want non-nil")
	}
}

func TestActivate_FailsIfCannotDialSock(t *testing.T) {
	setup(t)
	testutil.MustCreateEmpty("sock")
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err == nil {
		t.Errorf("got error nil, want non-nil")
	}
}
