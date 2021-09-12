package daemon

import (
	"io"
	"syscall"
	"testing"

	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/store/storetest"
	"src.elv.sh/pkg/testutil"
)

func TestProgram_ServesClientRequests(t *testing.T) {
	testutil.Umask(t, 0)
	testutil.InTempDir(t)

	// Set up server.
	serverDone := make(chan struct{})
	go func() {
		exit, _, stderr := Run(Program, "elvish", "-daemon", "-sock", "sock", "-db", "db")
		if exit != 0 {
			t.Logf("daemon exited with %v; stderr:\n%v", exit, stderr)
		}
		close(serverDone)
	}()
	defer func() { <-serverDone }()

	// Set up client.
	client, err := client.Activate(io.Discard,
		&daemondefs.SpawnConfig{SockPath: "sock", DbPath: "db", RunDir: "."})
	if err != nil {
		close(serverDone)
		t.Fatal("failed to activate client: ", err)
	}
	defer client.Close()

	// Test server state requests.
	gotVersion, err := client.Version()
	if gotVersion != api.Version || err != nil {
		t.Errorf(".Version() -> (%v, %v), want (%v, nil)", gotVersion, err, api.Version)
	}

	gotPid, err := client.Pid()
	wantPid := syscall.Getpid()
	if gotPid != wantPid || err != nil {
		t.Errorf(".Pid() -> (%v, %v), want (%v, nil)", gotPid, err, wantPid)
	}

	// Test store requests.
	storetest.TestCmd(t, client)
	storetest.TestDir(t, client)
	storetest.TestSharedVar(t, client)
}

func TestProgram_BadCLI(t *testing.T) {
	Test(t, Program,
		ThatElvish().
			ExitsWith(2).
			WritesStderr("internal error: no suitable subprogram\n"),

		ThatElvish("-daemon", "x").
			ExitsWith(2).
			WritesStderrContaining("arguments are not allowed with -daemon"),
	)
}
