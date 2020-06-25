package shell

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/parse"
)

// ScriptConfig keeps configuration for the script mode.
type ScriptConfig struct {
	SpawnDaemon bool
	Paths       Paths

	Cmd         bool
	CompileOnly bool
	JSON        bool
}

// Script executes a shell script.
func Script(fds [3]*os.File, args []string, cfg *ScriptConfig) int {
	ev, cleanup := setupShell(fds, cfg.Paths, cfg.SpawnDaemon)
	defer cleanup()

	arg0 := args[0]
	ev.SetArgs(args[1:])

	var name, code string
	if cfg.Cmd {
		name = "code from -c"
		code = arg0
	} else {
		var err error
		name, err = filepath.Abs(arg0)
		if err != nil {
			fmt.Fprintf(fds[2],
				"cannot get full path of script %q: %v\n", arg0, err)
			return 2
		}
		code, err = readFileUTF8(name)
		if err != nil {
			fmt.Fprintf(fds[2], "cannot read script %q: %v\n", name, err)
			return 2
		}
	}

	op, err := ev.ParseAndCompile(
		parse.Source{Name: name, Code: code, IsFile: true}, fds[2])
	if err != nil {
		if cfg.CompileOnly && cfg.JSON {
			fmt.Fprintf(fds[1], "%s\n", errorToJSON(err))
		} else {
			diag.ShowError(fds[2], err)
		}
		return 2
	}
	if cfg.CompileOnly {
		return 0
	}

	err = evalInTTY(ev, op, fds)
	if err != nil {
		diag.ShowError(fds[2], err)
		return 2
	}
	return 0
}

var errSourceNotUTF8 = errors.New("source is not UTF-8")

func readFileUTF8(fname string) (string, error) {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(bytes) {
		return "", errSourceNotUTF8
	}
	return string(bytes), nil
}
