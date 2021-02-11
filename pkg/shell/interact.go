package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/xiaq/persistent/hashmap"
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/sys"
)

// InteractiveRescueShell determines whether a panic results in a rescue shell
// being launched. It should be set to false by interactive mode unit tests.
var interactiveRescueShell bool = true

// InteractConfig keeps configuration for the interactive mode.
type InteractConfig struct {
	SpawnDaemon bool
	Paths       Paths
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
	ev, cleanup := setupShell(fds, cfg.Paths, cfg.SpawnDaemon)
	defer cleanup()

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
	if cfg.Paths.Rc != "" {
		err := sourceRC(fds, ev, cfg.Paths.Rc)
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

		err = evalInTTY(ev, fds,
			parse.Source{Name: fmt.Sprintf("[tty %v]", cmdNum), Code: line})
		term.Sanitize(fds[0], fds[2])
		if err != nil {
			diag.ShowError(fds[2], err)
		}
	}
}

func sourceRC(fds [3]*os.File, ev *eval.Evaler, rcPath string) error {
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
	err = evalInTTY(ev, fds, parse.Source{Name: absPath, Code: code, IsFile: true})
	if err != nil {
		return err
	}
	extraGlobal := extractExports(ev.Global(), fds[2])
	if extraGlobal != nil {
		ev.AddGlobal(extraGlobal)
	}
	return nil
}

const exportsVarName = "-exports-"

// If the namespace contains a variable named exportsVarName, extract its values
// into a namespace.
func extractExports(ns *eval.Ns, stderr io.Writer) *eval.Ns {
	value, ok := ns.Index(exportsVarName)
	if !ok {
		return nil
	}
	if prog.DeprecationLevel >= 15 {
		fmt.Fprintln(stderr,
			"the $-exports- mechanism is deprecated; use edit:add-vars instead.")
	}
	exports, ok := value.(hashmap.Map)
	if !ok {
		fmt.Fprintf(stderr, "$%s is not map, ignored\n", exportsVarName)
		return nil
	}
	nb := eval.NsBuilder{}
	for it := exports.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		name, ok := k.(string)
		if !ok {
			fmt.Fprintf(stderr, "$%s[%s] is not string, ignored\n",
				exportsVarName, vals.Repr(k, vals.NoPretty))
			continue
		}
		if ns.HasName(name) {
			fmt.Fprintf(stderr, "$%s already exists, ignored $%s[%s]\n",
				name, exportsVarName, name)
			continue
		}
		nb.Add(name, vars.FromInit(v))
	}
	return nb.Ns()
}
