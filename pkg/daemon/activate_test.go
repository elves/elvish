package daemon

import (
	"io"
	"net"
	"os"
	"runtime"
	"testing"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/testutil"
)

func TestActivate_ConnectsToExistingServer(t *testing.T) {
	setup(t)
	startServer(t, cli("sock", "db"))
	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
}

func TestActivate_SpawnsNewServer(t *testing.T) {
	activated := 0
	setupForActivate(t, func(name string, argv []string, attr *os.ProcAttr) error {
		startServer(t, argv)
		activated++
		return nil
	})

	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
	if activated != 1 {
		t.Errorf("got activated %v times, want 1", activated)
	}
}

func TestActivate_RemovesHangingSocketAndSpawnsNewServer(t *testing.T) {
	activated := 0
	setupForActivate(t, func(name string, argv []string, attr *os.ProcAttr) error {
		startServer(t, argv)
		activated++
		return nil
	})
	makeHangingUNIXSocket(t, "sock")

	_, err := Activate(io.Discard,
		&daemondefs.SpawnConfig{DbPath: "db", SockPath: "sock", RunDir: "."})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
	if activated != 1 {
		t.Errorf("got activated %v times, want 1", activated)
	}
}

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

func setupForActivate(t *testing.T, f func(string, []string, *os.ProcAttr) error) {
	setup(t)

	saveStartProcess := startProcess
	t.Cleanup(func() { startProcess = saveStartProcess })
	startProcess = f

	saveDaemonSpawnTimeout := daemonSpawnTimeout
	t.Cleanup(func() { daemonSpawnTimeout = saveDaemonSpawnTimeout })
	daemonSpawnTimeout = testutil.ScaledMs(1000)
}

func makeHangingUNIXSocket(t *testing.T, path string) {
	t.Helper()

	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	// We need to call l.Close() to make the socket hang, but that will
	// helpfully remove the socket file. Work around this by renaming the socket
	// file.
	err = os.Rename(path, path+".save")
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	err = os.Rename(path+".save", path)
	if err != nil {
		t.Fatal(err)
	}
}
