package shell_test

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
	"src.elv.sh/pkg/testutil"
)

//go:embed *.elvts
var transcripts embed.FS

var sigCHLDName = ""

func TestTranscripts(t *testing.T) {
	evaltest.TestTranscriptsInFS(t, transcripts,
		"elvish-in-global", func(ev *eval.Evaler) {
			ev.ExtendGlobal(eval.BuildNs().
				AddGoFn("elvish", programAsGoFn(&shell.Program{})))
		},
		"elvish-with-activate-daemon-in-global", func(t *testing.T, ev *eval.Evaler) {
			p := &shell.Program{ActivateDaemon: inProcessActivateFunc(t)}
			ev.ExtendGlobal(eval.BuildNs().AddGoFn("elvish", programAsGoFn(p)))
		},
		"elvish-with-bad-activate-daemon-in-global", func(ev *eval.Evaler) {
			p := &shell.Program{
				ActivateDaemon: func(io.Writer, *daemondefs.SpawnConfig) (daemondefs.Client, error) {
					return nil, errors.New("fake error")
				},
			}
			ev.ExtendGlobal(eval.BuildNs().AddGoFn("elvish", programAsGoFn(p)))
		},
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

type programOpts struct {
	CheckStdoutContains string
	CheckStderrContains string
}

func (programOpts) SetDefaultOptions() {}

// Converts a [prog.Program] to a Go-implemented Elvish function.
//
// Stdin of the program is connected to the stdin of the function.
//
// Stdout of the program is usually written unchanged to the stdout of the
// function, except when:
//
//   - If the output has no trailing newline, " (no EOL)\n" is appended.
//   - If &check-stdout-contains is supplied, stdout is suppressed. Instead, a
//     tag "[stdout contains foo]" is shown, followed by "true" or "false".
//
// Stderr of the program is written to the stderr of the function with a
// [stderr] prefix, with similar treatment for missing trailing EOL and
// &check-stderr-contains.
//
// If the program exits with a non-zero return value, a line "[exit] $i" is
// written to stderr.
func programAsGoFn(p prog.Program) any {
	return func(fm *eval.Frame, opts programOpts, args ...string) {
		r1, w1 := must.OK2(os.Pipe())
		r2, w2 := must.OK2(os.Pipe())
		args = append([]string{"elvish"}, args...)
		exit := prog.Run([3]*os.File{fm.InputFile(), w1, w2}, args, p)
		w1.Close()
		w2.Close()

		outFile := fm.ByteOutput()
		stdout := string(must.OK1(io.ReadAll(r1)))
		if s := opts.CheckStdoutContains; s != "" {
			fmt.Fprintf(outFile,
				"[stdout contains %q] %t\n", s, strings.Contains(stdout, s))
		} else {
			outFile.WriteString(lines("", stdout))
		}

		errFile := fm.ErrorFile()
		stderr := string(must.OK1(io.ReadAll(r2)))
		if s := opts.CheckStderrContains; s != "" {
			fmt.Fprintf(errFile,
				"[stderr contains %q] %t\n", s, strings.Contains(stderr, s))
		} else {
			errFile.WriteString(lines("[stderr] ", stderr))
		}

		if exit != 0 {
			fmt.Fprintf(errFile, "[exit] %d\n", exit)
		}
	}
}

// Splits data into lines, adding prefix to each line and appending " (no EOL)"
// if data doesn't end in a newline.
func lines(prefix, data string) string {
	var sb strings.Builder
	for len(data) > 0 {
		sb.WriteString(prefix)
		i := strings.IndexByte(data, '\n')
		if i == -1 {
			sb.WriteString(data + " (no EOL)\n")
			break
		} else {
			sb.WriteString(data[:i+1])
			data = data[i+1:]
		}
	}
	return sb.String()
}
