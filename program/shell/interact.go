package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/cliedit"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/sys"
	"github.com/xiaq/persistent/hashmap"
)

func interact(ev *eval.Evaler, dataDir string, norc bool) {
	// Build Editor.
	var ed editor
	if sys.IsATTY(os.Stdin) {
		newed := cliedit.NewEditor(cli.StdTTY, ev, ev.DaemonClient)
		ev.Global.AddNs("edit", newed.Ns())
		ed = newed
	} else {
		ed = newMinEditor(os.Stdin, os.Stderr)
	}
	defer ed.Close()

	// Source rc.elv.
	if !norc && dataDir != "" {
		err := sourceRC(ev, dataDir)
		if err != nil {
			diag.PPrintError(err)
		}
	}

	term.Sanitize(os.Stdin, os.Stderr)

	// Build readLine function.
	readLine := ed.ReadLine

	cooldown := time.Second
	usingBasic := false
	cmdNum := 0

	for {
		cmdNum++

		line, err := readLine()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Editor error:", err)
			if !usingBasic {
				fmt.Println("Falling back to basic line editor")
				readLine = basicReadLine
				usingBasic = true
			} else {
				fmt.Println("Don't know what to do, pid is", os.Getpid())
				fmt.Println("Restarting editor in", cooldown)
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
		term.Sanitize(os.Stdin, os.Stderr)
		if err != nil {
			diag.PPrintError(err)
		}
	}
}

func sourceRC(ev *eval.Evaler, dataDir string) error {
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
	extractExports(ev.Global, os.Stderr)
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

func basicReadLine() (string, error) {
	stdin := bufio.NewReaderSize(os.Stdin, 0)
	return stdin.ReadString('\n')
}
