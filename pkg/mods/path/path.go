// Package path provides functions for manipulating filesystem path names.
package path

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vars"
)

// Ns is the namespace for the path: module.
var Ns = eval.BuildNsNamed("path").
	AddVars(map[string]vars.Var{
		"dev-null":       vars.NewReadOnly(os.DevNull),
		"dev-tty":        vars.NewReadOnly(devTty),
		"list-separator": vars.NewReadOnly(string(filepath.ListSeparator)),
		"separator":      vars.NewReadOnly(string(filepath.Separator)),
	}).
	AddGoFns(map[string]any{
		"abs":           filepath.Abs,
		"base":          filepath.Base,
		"clean":         filepath.Clean,
		"dir":           filepath.Dir,
		"ext":           filepath.Ext,
		"eval-symlinks": filepath.EvalSymlinks,
		"is-abs":        filepath.IsAbs,
		"is-dir":        isDir,
		"is-regular":    isRegular,
		"join":          filepath.Join,
		"remove":        remove,
		"temp-dir":      tempDir,
		"temp-file":     tempFile,
	}).Ns()

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

type isOpts struct{ FollowSymlink bool }

func (opts *isOpts) SetDefaultOptions() {}

func isDir(opts isOpts, path string) bool {
	var fi os.FileInfo
	var err error
	if opts.FollowSymlink {
		fi, err = os.Stat(path)
	} else {
		fi, err = os.Lstat(path)
	}
	return err == nil && fi.Mode().IsDir()
}

func isRegular(opts isOpts, path string) bool {
	var fi os.FileInfo
	var err error
	if opts.FollowSymlink {
		fi, err = os.Stat(path)
	} else {
		fi, err = os.Lstat(path)
	}
	return err == nil && fi.Mode().IsRegular()
}

type mktempOpt struct{ Dir string }

func (o *mktempOpt) SetDefaultOptions() {}

func tempDir(opts mktempOpt, args ...string) (string, error) {
	var pattern string
	switch len(args) {
	case 0:
		pattern = "elvish-*"
	case 1:
		pattern = args[0]
	default:
		return "", errs.ArityMismatch{What: "arguments",
			ValidLow: 0, ValidHigh: 1, Actual: len(args)}
	}

	return os.MkdirTemp(opts.Dir, pattern)
}

func tempFile(opts mktempOpt, args ...string) (*os.File, error) {
	var pattern string
	switch len(args) {
	case 0:
		pattern = "elvish-*"
	case 1:
		pattern = args[0]
	default:
		return nil, errs.ArityMismatch{What: "arguments",
			ValidLow: 0, ValidHigh: 1, Actual: len(args)}
	}

	return os.CreateTemp(opts.Dir, pattern)
}

type rmOpts struct {
	IgnoreMissing bool
	Recursive     bool
}

func (opts *rmOpts) SetDefaultOptions() {}

// remove deletes filesystem paths.
func remove(opts rmOpts, args ...string) error {
	var returnErr error
	for _, path := range args {
		err := recursiveRemove(path, opts.Recursive, opts.IgnoreMissing)
		returnErr = errutil.Multi(returnErr, err)
	}
	return returnErr
}

// recursiveRemove deletes a filesystem path. It optimistically hopes that any
// path that refers to a non-directory or an empty directory. If a directory is
// not empty, and the `recursive` option is true, it will attempt to do a
// depth-first removal of the path.
func recursiveRemove(path string, recursive bool, ignoreMissing bool) error {
	err := os.Remove(path)
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		if ignoreMissing {
			return nil
		}
	} else if isDirNotEmpty(err.(*fs.PathError).Unwrap()) {
		if !recursive {
			return err
		}
		dirEntries, suberr := os.ReadDir(path)
		if suberr != nil {
			return errutil.Multi(err, suberr)
		}
		err = nil
		for _, f := range dirEntries {
			path := filepath.Join(path, f.Name())
			suberr := recursiveRemove(path, recursive, ignoreMissing)
			err = errutil.Multi(err, suberr)
		}
		suberr = os.Remove(path)
		return errutil.Multi(err, suberr)
	}
	return err
}
