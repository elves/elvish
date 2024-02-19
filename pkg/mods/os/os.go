// Package os exposes functionality from Go's os package as an Elvish module.
package os

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

// Ns is the Elvish namespace for this module.
var Ns = eval.BuildNsNamed("os").
	AddVars(map[string]vars.Var{
		"dev-null": vars.NewReadOnly(os.DevNull),
		"dev-tty":  vars.NewReadOnly(DevTTY),
	}).
	AddGoFns(map[string]any{
		"-is-exist":     isExist,
		"-is-not-exist": isNotExist,

		// File CRUD.
		"mkdir":      mkdir,
		"mkdir-all":  mkdirAll,
		"symlink":    os.Symlink,
		"remove":     remove,
		"remove-all": removeAll,
		"rename":     os.Rename,
		"chmod":      chmod,

		// File query.
		"stat":       stat,
		"exists":     exists,
		"is-dir":     IsDir,
		"is-regular": IsRegular,

		"eval-symlinks": filepath.EvalSymlinks,

		// Temp file/dir.
		"temp-dir":  TempDir,
		"temp-file": TempFile,
	}).Ns()

// Wraps [os.IsNotExist] to operate on Exception values.
func isExist(e eval.Exception) bool {
	return os.IsExist(e.Reason())
}

// Wraps [os.IsNotExist] to operate on Exception values.
func isNotExist(e eval.Exception) bool {
	return os.IsNotExist(e.Reason())
}

type mkdirOpts struct{ Perm int }

func (opts *mkdirOpts) SetDefaultOptions() { opts.Perm = 0755 }

func mkdir(opts mkdirOpts, path string) error {
	return os.Mkdir(path, os.FileMode(opts.Perm))
}

func mkdirAll(opts mkdirOpts, path string) error {
	return os.MkdirAll(path, os.FileMode(opts.Perm))
}

// ErrEmptyPath is thrown by remove and remove-all when given an empty path.
var ErrEmptyPath = errs.BadValue{
	What: "path", Valid: "non-empty string", Actual: "empty string"}

// Wraps [os.Remove] to reject empty paths.
func remove(path string) error {
	if path == "" {
		return ErrEmptyPath
	}
	return os.Remove(path)
}

// Wraps [os.RemoveAll] to reject empty paths, and resolve relative paths to
// absolute paths first. The latter is necessary since the working directory
// could be changed while [os.RemoveAll] is running.
func removeAll(path string) error {
	if path == "" {
		return ErrEmptyPath
	}
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		path = absPath
	}
	return os.RemoveAll(path)
}

type chmodOpts struct {
	SpecialModes any
}

func (*chmodOpts) SetDefaultOptions() {}

func chmod(opts chmodOpts, perm int, path string) error {
	if perm < 0 || perm > 0x777 {
		return errs.OutOfRange{What: "permission bits",
			ValidLow: "0", ValidHigh: "0o777", Actual: strconv.Itoa(perm)}
	}
	mode := fs.FileMode(perm)
	if opts.SpecialModes != nil {
		special, err := specialModesFromIterable(opts.SpecialModes)
		if err != nil {
			return err
		}
		mode |= special
	}
	return os.Chmod(path, mode)
}

type statOpts struct{ FollowSymlink bool }

func (opts *statOpts) SetDefaultOptions() {}

func stat(opts statOpts, path string) (vals.Map, error) {
	fi, err := statOrLstat(path, opts.FollowSymlink)
	if err != nil {
		return nil, err
	}
	return statMap(fi), nil
}

func exists(opts statOpts, path string) bool {
	_, err := statOrLstat(path, opts.FollowSymlink)
	return err == nil
}

// IsDir is exported so that the implementation may be shared by the path:
// module.
func IsDir(opts statOpts, path string) bool {
	fi, err := statOrLstat(path, opts.FollowSymlink)
	return err == nil && fi.Mode().IsDir()
}

// IsRegular is exported so that the implementation may be shared by the path:
// module.
func IsRegular(opts statOpts, path string) bool {
	fi, err := statOrLstat(path, opts.FollowSymlink)
	return err == nil && fi.Mode().IsRegular()
}

func statOrLstat(path string, followSymlink bool) (os.FileInfo, error) {
	if followSymlink {
		return os.Stat(path)
	} else {
		return os.Lstat(path)
	}
}

type mktempOpt struct{ Dir string }

func (o *mktempOpt) SetDefaultOptions() {}

// TempDir is exported so that the implementation may be shared by the path:
// module.
func TempDir(opts mktempOpt, args ...string) (string, error) {
	pattern, err := optionalTempPattern(args)
	if err != nil {
		return "", err
	}
	return os.MkdirTemp(opts.Dir, pattern)
}

// TempFile is exported so that the implementation may be shared by the path:
// module.
func TempFile(opts mktempOpt, args ...string) (*os.File, error) {
	pattern, err := optionalTempPattern(args)
	if err != nil {
		return nil, err
	}
	return os.CreateTemp(opts.Dir, pattern)
}

func optionalTempPattern(args []string) (string, error) {
	switch len(args) {
	case 0:
		return "elvish-*", nil
	case 1:
		return args[0], nil
	default:
		return "", errs.ArityMismatch{What: "arguments",
			ValidLow: 0, ValidHigh: 1, Actual: len(args)}
	}
}
