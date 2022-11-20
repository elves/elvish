package eval

import (
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/fsutil"
)

// Filesystem commands.

func init() {
	addBuiltinFns(map[string]any{
		// Directory
		"cd": cd,

		// Path
		"tilde-abbr": tildeAbbr,
	})
}

func cd(fm *Frame, args ...string) error {
	var dir string
	switch len(args) {
	case 0:
		var err error
		dir, err = getHome("")
		if err != nil {
			return err
		}
	case 1:
		dir = args[0]
	default:
		return errs.ArityMismatch{What: "arguments", ValidLow: 0, ValidHigh: 1, Actual: len(args)}
	}

	return fm.Evaler.Chdir(dir)
}

func tildeAbbr(path string) string {
	return fsutil.TildeAbbr(path)
}
