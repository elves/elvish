package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/daemon/daemondefs"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/mods/daemon"
	"src.elv.sh/pkg/mods/store"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/strutil"
	"src.elv.sh/pkg/sys"
	"src.elv.sh/pkg/ui"
)

// InteractiveRescueShell determines whether a panic results in a rescue shell
// being launched. It should be set to false by interactive mode unit tests.
var interactiveRescueShell bool = true

// Configuration for the interactive mode.
type interactCfg struct {
	RC string

	ActivateDaemon daemondefs.ActivateFunc
	SpawnConfig    *daemondefs.SpawnConfig
}

// Interface satisfied by the line editor. Used for swapping out the editor with
// minEditor when necessary.
type editor interface {
	ReadCode() (string, error)
	RunAfterCommandHooks(src parse.Source, duration float64, err error)
}

// Runs an interactive shell session.
func interact(ev *eval.Evaler, fds [3]*os.File, cfg *interactCfg) {
	if interactiveRescueShell {
		defer handlePanic()
	}

	var daemonClient daemondefs.Client
	if cfg.ActivateDaemon != nil && cfg.SpawnConfig != nil {
		// TODO(xiaq): Connect to daemon and install daemon module
		// asynchronously.
		cl, err := cfg.ActivateDaemon(fds[2], cfg.SpawnConfig)
		if err != nil {
			fmt.Fprintln(fds[2], "Cannot connect to daemon:", err)
			fmt.Fprintln(fds[2], "Daemon-related functions will likely not work.")
		}
		if cl != nil {
			// Even if error is not nil, we install daemon-related
			// functionalities anyway. Daemon may eventually come online and
			// become functional.
			daemonClient = cl
			ev.PreExitHooks = append(ev.PreExitHooks, func() { cl.Close() })
			ev.AddModule("store", store.Ns(cl))
			ev.AddModule("daemon", daemon.Ns(cl))
		}
	}

	// Build Editor.
	var ed editor
	if sys.IsATTY(fds[0].Fd()) {
		restoreTTY := term.SetupForTUIOnce(fds[0], fds[1])
		defer restoreTTY()
		newed := edit.NewEditor(cli.NewTTY(fds[0], fds[2]), ev, daemonClient)
		ev.ExtendBuiltin(eval.BuildNs().AddNs("edit", newed))
		ev.BgJobNotify = func(s string) { newed.Notify(ui.T(s)) }
		ed = newed
	} else {
		ed = newMinEditor(fds[0], fds[2])
	}

	// Source rc.elv.
	if cfg.RC != "" {
		err := sourceRC(fds, ev, ed, cfg.RC)
		if err != nil {
			diag.ShowError(fds[2], err)
		}
	}

	cooldown := time.Second
	cmdNum := 0

	for {
		cmdNum++

		line, err := ed.ReadCode()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(fds[2], "Editor error:", err)
			if _, isMinEditor := ed.(*minEditor); !isMinEditor {
				fmt.Fprintln(fds[2], "Falling back to basic line editor")
				ed = newMinEditor(fds[0], fds[2])
			} else {
				fmt.Fprintln(fds[2], "Don't know what to do, pid is", os.Getpid())
				fmt.Fprintln(fds[2], "Restarting editor in", cooldown)
				time.Sleep(cooldown)
				if cooldown < time.Minute {
					cooldown *= 2
				}
			}
			continue
		}

		// No error; reset cooldown.
		cooldown = time.Second

		// Execute the command line only if it is not entirely whitespace. This keeps side-effects,
		// such as executing `$edit:after-command` hooks, from occurring when we didn't actually
		// evaluate any code entered by the user.
		if strings.TrimSpace(line) == "" {
			continue
		}
		err = evalInTTY(fds, ev, ed,
			parse.Source{Name: fmt.Sprintf("[tty %v]", cmdNum), Code: line})
		if err != nil {
			diag.ShowError(fds[2], err)
		}
	}
}

// Interactive mode panic handler.
func handlePanic() {
	r := recover()
	if r != nil {
		println()
		print(sys.DumpStack())
		println()
		fmt.Println(r)
		println("\nExecing recovery shell /bin/sh")
		syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	}
}

func sourceRC(fds [3]*os.File, ev *eval.Evaler, ed editor, rcPath string) error {
	absPath, err := filepath.Abs(rcPath)
	if err != nil {
		return fmt.Errorf("cannot get full path of rc.elv: %v", err)
	}
	code, err := readFileUTF8(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return evalInTTY(fds, ev, ed, parse.Source{Name: absPath, Code: code, IsFile: true})
}

type minEditor struct {
	in  *bufio.Reader
	out io.Writer
}

func newMinEditor(in, out *os.File) *minEditor {
	return &minEditor{bufio.NewReader(in), out}
}

func (ed *minEditor) RunAfterCommandHooks(src parse.Source, duration float64, err error) {
	// no-op; minEditor doesn't support this hook.
}

func (ed *minEditor) ReadCode() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "?"
	}
	fmt.Fprintf(ed.out, "%s> ", wd)
	line, err := ed.in.ReadString('\n')
	return strutil.ChopLineEnding(line), err
}
