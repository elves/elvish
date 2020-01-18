package eval

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/store"
	"github.com/elves/elvish/pkg/util"
)

// Filesystem.

// ErrStoreNotConnected is thrown by dir-history when the store is not connected.
var ErrStoreNotConnected = errors.New("store not connected")

//elvdoc:fn cd
//
// ```elvish
// cd $dirname
// ```
//
// Change directory.
//
// Note that Elvish's `cd` does not support `cd -`.

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

// TODO(xiaq): Document eval-symlinks.

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

// TODO(xiaq): Document -is-dir.

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

func cd(fm *Frame, args ...string) error {
	var dir string
	switch len(args) {
	case 0:
		var err error
		dir, err = util.GetHome("")
		if err != nil {
			return err
		}
	case 1:
		dir = args[0]
	default:
		return ErrArgs
	}

	return fm.Chdir(dir)
}

type dirHistoryEntry struct {
	Path  string  `json:"path"`
	Score float64 `json:"score"`
}

func (dirHistoryEntry) IsStructMap(vals.StructMapMarker) {}

func dirs(fm *Frame) error {
	if fm.DaemonClient == nil {
		return ErrStoreNotConnected
	}
	dirs, err := fm.DaemonClient.Dirs(store.NoBlacklist)
	if err != nil {
		return err
	}
	out := fm.ports[1].Chan
	for _, dir := range dirs {
		out <- dirHistoryEntry{dir.Path, dir.Score}
	}
	return nil
}

func tildeAbbr(path string) string {
	return util.TildeAbbr(path)
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsDir()
}
