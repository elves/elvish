package daemon

import (
	"os"
	"syscall"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/daemon/internal/api"
	"src.elv.sh/pkg/must"
	. "src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/store/storetest"
	"src.elv.sh/pkg/testutil"
)

func TestProgram_TerminatesIfCannotListen(t *testing.T) {
	setup(t)
	must.CreateEmpty("sock")

	Test(t, &Program{},
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
}

func TestProgram_StillServesIfCannotOpenDB(t *testing.T) {
	setup(t)
	must.WriteFile("db", "not a valid bolt database")
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
	startServerOpts(t, cli("sock", "db"), ServeOpts{Signals: sigCh})
	close(sigCh)
	// startServerSigCh will wait for server to terminate at cleanup
}

func TestProgram_QuitsOnSignalChannelWithClients(t *testing.T) {
	setup(t)
	sigCh := make(chan os.Signal)
	server := startServerOpts(t, cli("sock", "db"), ServeOpts{Signals: sigCh})
	client := startClient(t, "sock")
	close(sigCh)

	server.WaitQuit()
	_, err := client.Version()
	if err == nil {
		t.Errorf("client.Version() returns nil error, want non-nil")
	}
}

func TestProgram_BadCLI(t *testing.T) {
	Test(t, &Program{},
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

// Calls startServerOpts with a Signals channel that gets closed during cleanup.
func startServer(t *testing.T, args []string) server {
	t.Helper()
	sigCh := make(chan os.Signal)
	s := startServerOpts(t, args, ServeOpts{Signals: sigCh})
	// Cleanup functions added later are run earlier. This will be run before
	// the cleanup function added by startServerOpts that waits for the server
	// to terminate.
	t.Cleanup(func() { close(sigCh) })
	return s
}

// Start server with custom ServeOpts (opts.Ready is ignored). Makes sure that
// the server terminates during cleanup.
func startServerOpts(t *testing.T, args []string, opts ServeOpts) server {
	t.Helper()
	readyCh := make(chan struct{})
	opts.Ready = readyCh
	doneCh := make(chan serverResult)
	go func() {
		exit, stdout, stderr := Run(&Program{serveOpts: opts}, args...)
		doneCh <- serverResult{exit, stdout, stderr}
		close(doneCh)
	}()
	select {
	case <-readyCh:
	case <-time.After(testutil.Scaled(2 * time.Second)):
		t.Fatal("timed out waiting for daemon to start")
	}
	s := server{t, doneCh}
	t.Cleanup(func() { s.WaitQuit() })
	return s
}

type server struct {
	t  *testing.T
	ch <-chan serverResult
}

type serverResult struct {
	exit           int
	stdout, stderr string
}

func (s server) WaitQuit() (serverResult, bool) {
	s.t.Helper()
	select {
	case r := <-s.ch:
		return r, true
	case <-time.After(testutil.Scaled(2 * time.Second)):
		s.t.Error("timed out waiting for daemon to quit")
		return serverResult{}, false
	}
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
