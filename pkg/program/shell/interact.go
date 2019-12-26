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

func interact(fds [3]*os.File, ev *eval.Evaler, dataDir string, norc bool) {
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
	if !norc && dataDir != "" {
		err := sourceRC(fds[2], ev, dataDir)
		if err != nil {
			diag.PPrintError(err)
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

		err = ev.EvalSourceInTTY(eval.NewInteractiveSource(line))
		term.Sanitize(fds[0], fds[2])
		if err != nil {
			diag.PPrintError(err)
		}
	}
}

func sourceRC(stderr *os.File, ev *eval.Evaler, dataDir string) error {
	absPath, err := filepath.Abs(filepath.Join(dataDir, "rc.elv"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot get full path of rc.elv: %v", err)
	}
	code, err := readFileUTF8(absPath)
	err = ev.EvalSourceInTTY(eval.NewScriptSource("rc.elv", absPath, code))
	if err != nil {
		return err
	}
	extractExports(ev.Global, stderr)
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
