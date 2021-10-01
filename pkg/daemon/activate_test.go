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
	setup(t)
	// Build a path for which Lstat will return a non-nil err such that
	// os.IsNotExist(err) is false.
	badSockPath := ""
	if runtime.GOOS != "windows" {
		// POSIX lstat(2) returns ENOTDIR instead of ENOENT if a path prefix is
		// not a directory.
		testutil.MustCreateEmpty("not-dir")
		badSockPath = "not-dir/sock"
	} else {
		// Use a syntactically invalid drive letter on Windows.
		badSockPath = `CD:\sock`
	}
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: badSockPath, RunDir: "."})
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
