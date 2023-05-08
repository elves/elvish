// Package os exposes functionality from Go's os package as an Elvish module.
package os

import (
	_ "embed"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
)

// Ns is the Elvish namespace for this module.
var Ns = eval.BuildNsNamed("os").
	AddGoFns(map[string]any{
		// Thin wrappers of functions in the Go package.
		"-is-exist":     isExist,
		"-is-not-exist": isNotExist,
		"mkdir":         mkdir,
		"remove":        remove,
		"remove-all":    removeAll,
		// Higher-level utilities.
		"exists": exists,
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

type existsOpts struct{ FollowSymlink bool }

func (opts *existsOpts) SetDefaultOptions() {}

func exists(opts existsOpts, path string) bool {
	var err error
	if opts.FollowSymlink {
		_, err = os.Stat(path)
	} else {
		_, err = os.Lstat(path)
	}
	return err == nil
}
