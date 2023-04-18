// Package path provides functions for manipulating filesystem path names.
package os

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval"
)

// Ns is the namespace for the path: module.
var Ns = eval.BuildNsNamed("path").
	AddGoFns(map[string]any{
		"remove": remove,
	}).Ns()

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

type rmOpts struct {
	MissingOk bool
	Recursive bool
}

func (opts *rmOpts) SetDefaultOptions() {}

// remove deletes filesystem paths.
func remove(opts rmOpts, args ...string) error {
	var returnErr error
	for _, path := range args {
		err := recursiveRemove(path, opts.Recursive, opts.MissingOk)
		returnErr = errutil.Multi(returnErr, err)
	}
	return returnErr
}

// recursiveRemove deletes a filesystem path. It optimistically assumes that any
// path refers to a non-directory or an empty directory. If a directory is not
// empty, and the `recursive` option is true, it will attempt to do a
// depth-first removal of the path. This does not use the Go os.RemoveAll
// function because we want to include all paths that could not be removed in
// the error this can return. We also want to distinguish between a path not
// existing versus other errors and os.RemoveAll makes that harder. The logic is
// simpler if we only rely on os.Remove to disambiguate these cases.
func recursiveRemove(path string, recursive bool, MissingOk bool) error {
	err := os.Remove(path)
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		if MissingOk {
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
			suberr := recursiveRemove(path, recursive, MissingOk)
			err = errutil.Multi(err, suberr)
		}
		suberr = os.Remove(path)
		return errutil.Multi(err, suberr)
	}
	return err
}
