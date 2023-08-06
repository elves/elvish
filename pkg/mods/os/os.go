// Package os exposes functionality from Go's os package as an Elvish module.
package os

import (
	_ "embed"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
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

		"mkdir":      mkdir,
		"remove":     remove,
		"remove-all": removeAll,

		"eval-symlinks": filepath.EvalSymlinks,

		"exists":     exists,
		"is-dir":     IsDir,
		"is-regular": IsRegular,

		"temp-dir":  TempDir,
		"temp-file": TempFile,
	}).Ns()

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

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

type statPredOpts struct{ FollowSymlink bool }

func (opts *statPredOpts) SetDefaultOptions() {}

func exists(opts statPredOpts, path string) bool {
	_, err := stat(path, opts.FollowSymlink)
	return err == nil
}

// IsDir is exported so that the implementation may be shared by the path:
// module.
func IsDir(opts statPredOpts, path string) bool {
	fi, err := stat(path, opts.FollowSymlink)
	return err == nil && fi.Mode().IsDir()
}

// IsRegular is exported so that the implementation may be shared by the path:
// module.
func IsRegular(opts statPredOpts, path string) bool {
	fi, err := stat(path, opts.FollowSymlink)
	return err == nil && fi.Mode().IsRegular()
}

func stat(path string, followSymlink bool) (os.FileInfo, error) {
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
