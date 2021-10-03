package eval

import (
	"errors"

	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/store/storedefs"

	"src.elv.sh/pkg/eval/errs"
)

// Filesystem commands.

// ErrStoreNotConnected is thrown by dir-history when the store is not connected.
var ErrStoreNotConnected = errors.New("store not connected")

func init() {
	addBuiltinFns(map[string]interface{}{
		// Directory
		"cd":          cd,
		"dir-history": dirs,

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

//elvdoc:fn dir-history
//
// ```elvish
// dir-history
// ```
//
// Return a list containing the interactive directory history. Each element is a map with two keys:
// `path` and `score`. The list is sorted by descending score.
//
// Example:
//
// ```elvish-transcript
// ~> dir-history | take 1
// ▶ [&path=/Users/foo/.elvish &score=96.79928]
// ```
//
// @cf edit:command-history

type dirHistoryEntry struct {
	Path  string
	Score float64
}

func (dirHistoryEntry) IsStructMap() {}

func dirs(fm *Frame) error {
	daemon := fm.Evaler.DaemonClient()
	if daemon == nil {
		return ErrStoreNotConnected
	}
	dirs, err := daemon.Dirs(storedefs.NoBlacklist)
	if err != nil {
		return err
	}
	out := fm.ValueOutput()
	for _, dir := range dirs {
		err := out.Put(dirHistoryEntry{dir.Path, dir.Score})
		if err != nil {
			return err
		}
	}
	return nil
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
