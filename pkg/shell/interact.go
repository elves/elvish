package shell

import (
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
	daemonmod "src.elv.sh/pkg/eval/mods/daemon"
	"src.elv.sh/pkg/eval/mods/store"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/sys"
)

// InteractiveRescueShell determines whether a panic results in a rescue shell
// being launched. It should be set to false by interactive mode unit tests.
var interactiveRescueShell bool = true

// InteractConfig keeps configuration for the interactive mode.
type InteractConfig struct {
	Evaler *eval.Evaler
	RC     string

	ActivateDaemon daemondefs.ActivateFunc
	SpawnConfig    *daemondefs.SpawnConfig
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

// Interact runs an interactive shell session.
func Interact(fds [3]*os.File, cfg *InteractConfig) {
	if interactiveRescueShell {
		defer handlePanic()
	}
	ev := cfg.Evaler

	if cfg.ActivateDaemon != nil && cfg.SpawnConfig != nil {
		// TODO(xiaq): Connect to daemon and install daemon module
		// asynchronously.
		cl, err := cfg.ActivateDaemon(fds[2], cfg.SpawnConfig)
		if err != nil {
			fmt.Fprintln(fds[2], "Cannot connect to daemon:", err)
			fmt.Fprintln(fds[2], "Daemon-related functions will likely not work.")
		}
		defer func() {
			err := cl.Close()
			if err != nil {
				fmt.Fprintln(fds[2],
					"warning: failed to close connection to daemon:", err)
			}
		}()
		// Even if error is not nil, we install daemon-related functionalities
		// anyway. Daemon may eventually come online and become functional.
		ev.SetDaemonClient(cl)
		ev.AddModule("store", store.Ns(cl))
		ev.AddModule("daemon", daemonmod.Ns(cl))
	}

	// Build Editor.
	var ed editor
	if sys.IsATTY(fds[0]) {
		newed := edit.NewEditor(cli.NewTTY(fds[0], fds[2]), ev, ev.DaemonClient())
		ev.AddBuiltin(eval.NsBuilder{}.AddNs("edit", newed.Ns()).Ns())
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

	term.Sanitize(fds[0], fds[2])

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
		src := parse.Source{Name: fmt.Sprintf("[tty %v]", cmdNum), Code: line}
		duration, err := evalInTTY(ev, fds, src)
		ed.RunAfterCommandHooks(src, duration, err)
		term.Sanitize(fds[0], fds[2])
		if err != nil {
			diag.ShowError(fds[2], err)
		}
	}
}

func sourceRC(fds [3]*os.File, ev *eval.Evaler, ed eval.Editor, rcPath string) error {
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
	src := parse.Source{Name: absPath, Code: code, IsFile: true}
	duration, err := evalInTTY(ev, fds, src)
	ed.RunAfterCommandHooks(src, duration, err)
	return err
}
