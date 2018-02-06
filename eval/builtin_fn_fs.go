package eval

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

// Filesystem.

var ErrStoreNotConnected = errors.New("store not connected")

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
		dir = mustGetHome("")
	case 1:
		dir = args[0]
	default:
		return ErrArgs
	}

	return Chdir(dir, fm.DaemonClient)
}

func cdInner(dir string, fm *Frame) {
	maybeThrow(Chdir(dir, fm.DaemonClient))
}

var dirDescriptor = types.NewStructDescriptor("path", "score")

func newDirStruct(path string, score float64) *types.Struct {
	return types.NewStruct(dirDescriptor,
		[]interface{}{path, floatToElv(score)})
}

func dirs(ec *Frame) error {
	if ec.DaemonClient == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.DaemonClient.Dirs(storedefs.NoBlacklist)
	if err != nil {
		return err
	}
	out := ec.ports[1].Chan
	for _, dir := range dirs {
		out <- newDirStruct(dir.Path, dir.Score)
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
