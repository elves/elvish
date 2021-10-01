package daemon

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

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
	startServer(t, cli("sock", "db"))
	client := startClient(t, "sock")

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
	startServer(t, cli("sock", "db"))
	client := startClient(t, "sock")

	_, err := client.AddCmd("cmd")
	if err == nil {
		t.Errorf("got nil error, want non-nil")
	}
}

func TestProgram_QuitsOnSignalChannelWithNoClient(t *testing.T) {
	setup(t)
	sigCh := make(chan os.Signal)
	startServerSigCh(t, cli("sock", "db"), sigCh)
	close(sigCh)
	// startServerSigCh will wait for server to terminate at cleanup
}

func TestProgram_QuitsOnSignalChannelWithClients(t *testing.T) {
	setup(t)
	sigCh := make(chan os.Signal)
	doneCh := startServerSigCh(t, cli("sock", "db"), sigCh)
	client := startClient(t, "sock")
	close(sigCh)

	waitDone(t, doneCh)
	_, err := client.Version()
	if err == nil {
		t.Errorf("client.Version() returns nil error, want non-nil")
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

func startServer(t *testing.T, args []string) <-chan struct{} {
	sigCh := make(chan os.Signal)
	doneCh := startServerSigCh(t, args, sigCh)
	t.Cleanup(func() { close(sigCh) })
	return doneCh
}

func startServerSigCh(t *testing.T, args []string, sigCh <-chan os.Signal) <-chan struct{} {
	readyCh := make(chan struct{})
	doneCh := make(chan struct{})
	go func() {
		exit, stdout, stderr := Run(
			program{ServeChans{Ready: readyCh, Signal: sigCh}}, args...)
		if exit != 0 {
			fmt.Println("daemon exited with", exit)
			fmt.Print("stdout:\n", stdout)
			fmt.Print("stderr:\n", stderr)
		}
		close(doneCh)
	}()
	select {
	case <-readyCh:
	case <-time.After(testutil.ScaledMs(1000)):
		t.Fatal("timed out waiting for daemon to start")
	}
	t.Cleanup(func() { waitDone(t, doneCh) })
	return doneCh
}

func cli(sock, db string) []string {
	return []string{"elvish", "-daemon", "-sock", sock, "-db", db}
}

func startClient(t *testing.T, sock string) daemondefs.Client {
	cl := NewClient("sock")
	if _, err := cl.Version(); err != nil {
		t.Errorf("failed to start client: %v", err)
	}
	t.Cleanup(func() { cl.Close() })
	return cl
}

func waitDone(t *testing.T, doneCh <-chan struct{}) {
	select {
	case <-doneCh:
	case <-time.After(testutil.ScaledMs(1000)):
		t.Error("timed out waiting for daemon to quit")
	}
}
