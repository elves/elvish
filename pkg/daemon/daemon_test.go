package daemon

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon/client"
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
	startServer(t)
	client := client.NewClient("sock")

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
	startServer(t)
	client := client.NewClient("sock")

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

func startServer(t *testing.T) {
	t.Helper()

	readyCh := make(chan struct{})
	quitCh := make(chan interface{})
	doneCh := make(chan struct{})
	go func() {
		exit, stdout, stderr := Run(
			program{ServeChans{Ready: readyCh, Quit: quitCh}},
			"elvish", "-daemon", "-sock", "sock", "-db", "db")
		if exit != 0 {
			fmt.Println("daemon exited with", exit)
			fmt.Print("stdout:\n", stdout)
			fmt.Print("stderr:\n", stderr)
		}
		close(doneCh)
	}()
	select {
	case <-readyCh:
	case <-time.After(testutil.ScaledMs(100)):
		t.Fatal("timed out waiting for daemon to start")
	}
	t.Cleanup(func() {
		close(quitCh)
		<-doneCh
	})
}
