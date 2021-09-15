package daemon

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon/client"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/store/storetest"
	"src.elv.sh/pkg/testutil"
)

func TestProgram_TerminatesIfCannotListen(t *testing.T) {
	setup(t)
	testutil.MustCreateEmpty("sock")

	Test(t, Program,
		ThatElvish("-daemon", "-sock", "sock", "-db", "db").
			ExitsWith(2).
			WritesStdoutContaining("failed to listen on sock"),
	)
}

func TestProgram_ServesClientRequests(t *testing.T) {
	setup(t)
	client := startServerClientPair(t)

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

func TestProgram_StillServesIfCannotOpenDB(t *testing.T) {
	setup(t)
	testutil.MustWriteFile("db", "not a valid bolt database")
	client := startServerClientPair(t)

	_, err := client.AddCmd("cmd")
	if err == nil {
		t.Errorf("got nil error, want non-nil")
	}
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

func setup(t *testing.T) {
	testutil.Umask(t, 0)
	testutil.InTempDir(t)
}

func startServerClientPair(t *testing.T) daemondefs.Client {
	go startServer(t)
	client, err := startClient(t)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func startServer(t *testing.T) {
	exit, stdout, stderr := Run(Program, "elvish", "-daemon", "-sock", "sock", "-db", "db")
	if exit != 0 {
		fmt.Println("daemon exited with", exit)
		fmt.Print("stdout:\n", stdout)
		fmt.Print("stderr:\n", stderr)
	}
}

func startClient(t *testing.T) (daemondefs.Client, error) {
	client := client.NewClient("sock")
	t.Cleanup(func() { client.Close() })
	start := time.Now()
	timeout := testutil.ScaledMs(1000)
	for {
		client.ResetConn()
		_, err := client.Version()
		if err == nil {
			return client, nil
		}
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("Failed to connect after %v: %v", timeout, err)
		}
		time.Sleep(testutil.ScaledMs(10))
	}
}
