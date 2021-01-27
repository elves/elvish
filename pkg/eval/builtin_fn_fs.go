package eval

import (
	"errors"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/store"
)

// Filesystem commands.

// ErrStoreNotConnected is thrown by dir-history when the store is not connected.
var ErrStoreNotConnected = errors.New("store not connected")

//elvdoc:fn path-\*
//
// ```elvish
// path-abs $path
// path-base $path
// path-clean $path
// path-dir $path
// path-ext $path
// ```
//
// See [godoc of path/filepath](https://godoc.org/path/filepath). Go errors are
// turned into exceptions.
//
// These functions are deprecated. Use the equivalent functions in the
// [path:](path.html) module.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Directory
		"cd":          cd,
		"dir-history": dirs,

		// Path
		"path-abs":      filepath.Abs,
		"path-base":     filepath.Base,
		"path-clean":    filepath.Clean,
		"path-dir":      filepath.Dir,
		"path-ext":      filepath.Ext,
		"eval-symlinks": filepath.EvalSymlinks,
		"tilde-abbr":    tildeAbbr,

		// File types
		"-is-dir": isDir,
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
		return ErrArgs
	}

	return fm.Evaler.Chdir(dir)
}

//elvdoc:fn dir-history
//
// ```elvish
// dir-history
// ```
//
// Return a list containing the directory history. Each element is a map with two
// keys: `path` and `score`. The list is sorted by descending score.
//
// Example:
//
// ```elvish-transcript
// ~> dir-history | take 1
// ▶ [&path=/Users/foo/.elvish &score=96.79928]
// ```

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
	dirs, err := daemon.Dirs(store.NoBlacklist)
	if err != nil {
		return err
	}
	out := fm.OutputChan()
	for _, dir := range dirs {
		out <- dirHistoryEntry{dir.Path, dir.Score}
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

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsDir()
}
