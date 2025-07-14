package eval

import (
	"os"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/fsutil"
)

// Command and process control.

// TODO(xiaq): Document "fg".

func init() {
	addBuiltinFns(map[string]any{
		// Command resolution
		"external":        external,
		"has-external":    hasExternal,
		"search-external": searchExternal,

		// Process control
		"fg":   fg,
		"exec": execFn,
		"exit": exit,
	})
}

func external(cmd string) Callable {
	return NewExternalCmd(cmd)
}

func hasExternal(cmd string) bool {
	_, err := fsutil.SearchExecutable(cmd)
	return err == nil
}

func searchExternal(cmd string) (string, error) {
	return fsutil.SearchExecutable(cmd)
}

// Can be overridden in tests.
var osExit = os.Exit

func exit(fm *Frame, codes ...int) error {
	code := 0
	switch len(codes) {
	case 0:
	case 1:
		code = codes[0]
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: len(codes)}
	}

	fm.Evaler.PreExit()
	osExit(code)
	return nil
}
