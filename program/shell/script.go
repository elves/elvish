package shell

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"unicode/utf8"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// script evaluates a script. The returned error contains enough context and can
// be printed as-is (with util.PprintError).
func script(ev *eval.Evaler, args []string, cmd, compileOnly bool) error {
	arg0 := args[0]
	ev.SetArgs(args[1:])

	var name, path, code string
	if cmd {
		name = "code from -c"
		path = ""
		code = arg0
	} else {
		var err error
		name = arg0
		path, err = filepath.Abs(name)
		if err != nil {
			return fmt.Errorf("cannot get full path of script %q: %v", name, err)
		}
		code, err = readFileUTF8(path)
		if err != nil {
			return fmt.Errorf("cannot read script %q: %v", name, err)
		}
	}

	n, err := parse.Parse(name, code)
	if err != nil {
		return err
	}

	src := eval.NewScriptSource(name, path, code)
	op, err := ev.Compile(n, src)
	if err != nil {
		return err
	}
	if compileOnly {
		return nil
	}

	return ev.EvalWithStdPorts(op, src)
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
