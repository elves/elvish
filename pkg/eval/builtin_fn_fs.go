package eval

import (
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/fsutil"
)

// Filesystem commands.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Directory
		"cd": cd,

		// Path
		"tilde-abbr": tildeAbbr,
	})
}

//elvdoc:fn cd
//
// ```elvish
// cd $dirname
// ```
//
// Change directory. This affects the entire process; i.e., all threads
// whether running indirectly (e.g., prompt functions) or started explicitly
// by commands such as [`peach`](#peach).
//
// Note that Elvish's `cd` does not support `cd -`.
//
// @cf pwd

func cd(fm *Frame, args ...string) error {
	var dir string
	switch len(args) {
	case 0:
		var err error
		dir, err = fsutil.GetHome("")
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

//elvdoc:fn tilde-abbr
//
// ```elvish
// tilde-abbr $path
// ```
//
// If `$path` represents a path under the home directory, replace the home
// directory with `~`. Examples:
//
// ```elvish-transcript
// ~> echo $E:HOME
// /Users/foo
// ~> tilde-abbr /Users/foo
// ▶ '~'
// ~> tilde-abbr /Users/foobar
// ▶ /Users/foobar
// ~> tilde-abbr /Users/foo/a/b
// ▶ '~/a/b'
// ```

func tildeAbbr(path string) string {
	return fsutil.TildeAbbr(path)
}
