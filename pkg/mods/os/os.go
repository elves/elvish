// Package os exposes functionality from Go's os package as an Elvish module.
package os

import (
	_ "embed"
	"fmt"
	"io/fs"
	"math/big"
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

		"chmod":      chmod,
		"mkdir":      mkdir,
		"remove":     remove,
		"remove-all": removeAll,

		"eval-symlinks": filepath.EvalSymlinks,

		"stat":       stat,
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

const (
	// These are the publicly visible non-permission file mode bits. We map
	// these to the Go fs package equivalents. Other fs.FileMode bits (ignoring
	// the permission bits) are excluded from the mode value we make public.
	stickyBit = uint32(0o1000)
	setGidBit = uint32(0o2000)
	setUidBit = uint32(0o4000)
)

// convertModes does two things:
//
// 1) map various representations of file mode bits to a simple unsigned integer
// suitable for the `os.Chmod` function, and
//
// 2) map legacy POSIX non-permission special bits to the equivalent bits
// recognized by the `os.Chmod` function.
func convertModes(val any) (uint32, error) {
	var mode uint64
	var err error
	switch val := val.(type) {
	case string:
		// Assume the user provided an octal number without an explicit base
		// prefix. If that fails try to parse it as if it has an explicit prefix
		// or the value is base 10. This is consistent with the historical POSIX
		// chmod command.
		//
		// TODO: Decide if support for symbolic absolute and relative modes
		// should be added. Such as `=rw`, `go=`, or `go=u-w` (all from the
		// chmod man page on macOS). If the decision is not to support symbolic
		// modes then this comment should be replaced with a comment explaining
		// whey symbolic modes are not supported.
		if mode, err = strconv.ParseUint(val, 8, 32); err != nil {
			mode, err = strconv.ParseUint(val, 0, 32)
		}
	case int:
		mode = uint64(val)
	case *big.Int:
		if val.IsUint64() {
			mode = val.Uint64()
		} else {
			err = errs.Generic
		}
	default:
		err = errs.Generic
	}

	if err != nil || (mode&0o7777) != mode {
		return 0, errs.OutOfRange{
			What:      "mode (an integer)",
			ValidLow:  "0",
			ValidHigh: "0o7777",
			Actual:    fmt.Sprintf("%v", val),
		}
	}

	// We've validated mode is a 32 bit value via the range test above.
	// This block exists to map the non-permission mode special bits.
	filePerms := uint32(mode)
	if filePerms&stickyBit == stickyBit {
		filePerms = uint32(fs.ModeSticky) | (filePerms & ^stickyBit)
	}
	if filePerms&setGidBit == setGidBit {
		filePerms = uint32(fs.ModeSetgid) | (filePerms & ^setGidBit)
	}
	if filePerms&setUidBit == setUidBit {
		filePerms = uint32(fs.ModeSetuid) | (filePerms & ^setUidBit)
	}
	return filePerms, nil
}

// chmod modifies the mode (primarily the permissions) of a filesystem path.
func chmod(modes any, path string) error {
	fileMode, err := convertModes(modes)
	if err != nil {
		return err
	}

	newMode := fs.FileMode(fileMode)
	return os.Chmod(path, newMode)
}
