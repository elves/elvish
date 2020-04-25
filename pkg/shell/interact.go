package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/edit"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/sys"
	"github.com/xiaq/persistent/hashmap"
)

// InteractConfig keeps configuration for the interactive mode.
type InteractConfig struct {
	SpawnDaemon bool
	Paths       Paths
}

// Interact runs an interactive shell session.
func Interact(fds [3]*os.File, cfg *InteractConfig) {
	defer rescue()
	ev, cleanup := setupShell(fds, cfg.Paths, cfg.SpawnDaemon)
	defer cleanup()

	// Build Editor.
	var ed editor
	if sys.IsATTY(fds[0]) {
		newed := edit.NewEditor(cli.StdTTY, ev, ev.DaemonClient)
		ev.Builtin.AddNs("edit", newed.Ns())
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

		src := &eval.Source{Name: fmt.Sprintf("[tty %v]", cmdNum), Code: line}
		op, err := ev.ParseAndCompile(src)
		if err == nil {
			err = evalInTTY(ev, op, fds)
			term.Sanitize(fds[0], fds[2])
		}
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
	src := &eval.Source{Name: absPath, Code: code, IsFile: true}
	op, err := ev.ParseAndCompile(src)
	if err != nil {
		return err
	}
	err = evalInTTY(ev, op, fds)
	if err != nil {
		return err
	}
	extractExports(ev.Global, fds[2])
	return nil
}

const exportsVarName = "-exports-"

// If the namespace contains a variable named exportsVarName, extract its values
// into the namespace itself.
func extractExports(ns eval.Ns, stderr io.Writer) {
	if !ns.HasName(exportsVarName) {
		return
	}
	value := ns.PopName(exportsVarName).Get()
	exports, ok := value.(hashmap.Map)
	if !ok {
		fmt.Fprintf(stderr, "$%s is not map, ignored\n", exportsVarName)
		return
	}
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
		ns.Add(name, vars.FromInit(v))
	}
}
