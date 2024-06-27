package shell_test

import (
	"embed"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/prog/progtest"
	"src.elv.sh/pkg/shell"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

var sigCHLDName = ""

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"elvish-in-global", progtest.ElvishInGlobal(&shell.Program{}),
		"elvish-with-activate-daemon-in-global", progtest.ElvishInGlobal(
			&shell.Program{ActivateDaemon: inProcessActivateFunc(t)}),
		"elvish-with-bad-activate-daemon-in-global", progtest.ElvishInGlobal(
			&shell.Program{
				ActivateDaemon: func(io.Writer, *daemondefs.SpawnConfig) (daemondefs.Client, error) {
					return nil, errors.New("fake error")
				},
			}),
		"kill-wait-in-global", addGlobal("kill-wait",
			testutil.Scaled(10*time.Millisecond).String()),
		"sigchld-name-in-global", addGlobal("sigchld-name", sigCHLDName),
		"in-temp-home", func(t *testing.T) { testutil.InTempHome(t) },
	)
}

func inProcessActivateFunc(t *testing.T) daemondefs.ActivateFunc {
	return func(stderr io.Writer, cfg *daemondefs.SpawnConfig) (daemondefs.Client, error) {
		// Start an in-process daemon.
		//
		// Create the socket in a temporary directory. This is necessary because
		// we don't do enough mocking in the tests yet, and cfg.SockPath will
		// point to the socket used by real Elvish sessions.
		dir := testutil.TempDir(t)
		sockPath := filepath.Join(dir, "sock")
		sigCh := make(chan os.Signal)
		readyCh := make(chan struct{})
		daemonDone := make(chan struct{})
		go func() {
			// Unlike the socket path, we do honor cfg.DBPath; this is because
			// we run tests in a temporary HOME, so there's no risk of using the
			// DB of real Elvish sessions.
			daemon.Serve(sockPath, cfg.DbPath,
				daemon.ServeOpts{Ready: readyCh, Signals: sigCh})
			close(daemonDone)
		}()
		t.Cleanup(func() {
			close(sigCh)
			select {
			case <-daemonDone:
			case <-time.After(testutil.Scaled(2 * time.Second)):
				t.Errorf("timed out waiting for daemon to quit")
			}
		})
		select {
		case <-readyCh:
			// Do nothing
		case <-time.After(testutil.Scaled(2 * time.Second)):
			t.Fatalf("timed out waiting for daemon to start")
		}
		// Connect to it.
		return daemon.NewClient(sockPath), nil
	}
}

func addGlobal(name string, value any) func(ev *eval.Evaler) {
	return func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddVar(name, vars.NewReadOnly(value)))
	}
}
