// Package os exposes functionality from Go's os package as an Elvish module.
package os

import (
	_ "embed"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// fileInfo exposes metadata from the os.Stat/os.Lstat functions.
//
// We ignore a few fs.FileMode bits, such as fs.ModeAppend, because they aren't
// found on platforms currently supported by Elvish; e.g., Plan9. Many of the
// structure members are only meaningful on a subset of the platforms supported
// by Elvish. If not meaningful on a given platform it will have the zero value.
type fileInfo struct {
	Path          string
	AbsPath       string
	IsDir         bool
	IsRegular     bool
	IsSymlink     bool
	IsDevice      bool
	IsCharDevice  bool
	IsNamedPipe   bool
	IsSocket      bool
	Size          *big.Int
	Mode          *big.Int
	SymbolicMode  string
	Perms         *big.Int
	SymbolicPerms string
	MTime         vals.Time
	ATime         vals.Time
	BTime         vals.Time
	CTime         vals.Time
	Owner         string
	Group         string
	Uid           *big.Int
	Gid           *big.Int
	NumLinks      *big.Int
	Inode         *big.Int
	Device        *big.Int
	RawDevice     *big.Int
	BlockSize     *big.Int
	BlockCount    *big.Int
}

func (fileInfo) IsStructMap() {}

// Ns is the Elvish namespace for this module.
var Ns = eval.BuildNsNamed("os").
	AddGoFns(map[string]any{
		// Thin wrappers of functions in the Go package.
		"-is-exist":     isExist,
		"-is-not-exist": isNotExist,
		"mkdir":         mkdir,
		"remove":        remove,
		"remove-all":    removeAll,
		"stat":          stat,
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

const (
	// These are the publicly visible auxiliary permission file mode bits. We
	// map these to/from the Go fs package equivalents.
	publicStickyBit = uint64(0o1000)
	publicSetgidBit = uint64(0o2000)
	publicSetuidBit = uint64(0o4000)
)

type statOpts struct{ FollowSymlink bool }

func (opts *statOpts) SetDefaultOptions() {}

// stat outputs a psuedo-map containing metadata about each path.
func stat(fm *eval.Frame, opts statOpts, paths ...string) error {
	var returnErr error
	out := fm.ValueOutput()
	for _, path := range paths {
		var err error
		var info os.FileInfo
		if opts.FollowSymlink {
			info, err = os.Stat(path)
		} else {
			info, err = os.Lstat(path)
		}
		if err != nil {
			returnErr = errutil.Multi(returnErr, err)
		} else {
			out.Put(pathMetadata(path, info))
		}
	}
	return returnErr
}

// publicPerms exposes only the public permission bits of the file mode to make
// it easier for users to deal with the file permissions (which is a subset of
// the file mode bits). This maps three Go internal permission bits (setuid,
// setgid, sticky) to well known, legacy Unix, public bits to make is easier for
// users to deal with the file permissions and to use the permissions in Elvish
// commands such as `path:chmod`.
func publicPerms(info fs.FileInfo) (uint64, string) {
	perms := info.Mode() & (fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky)
	symbolicPerms := perms.String()
	numPerms := uint64(perms & fs.ModePerm)
	if perms&fs.ModeSetuid == fs.ModeSetuid {
		numPerms = numPerms | publicSetuidBit
	}
	if perms&fs.ModeSetgid == fs.ModeSetgid {
		numPerms = numPerms | publicSetgidBit
	}
	if perms&fs.ModeSticky == fs.ModeSticky {
		numPerms = numPerms | publicStickyBit
	}
	return numPerms, symbolicPerms
}
