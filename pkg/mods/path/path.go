// Package path provides functions for manipulating filesystem path names.
package path

import (
	_ "embed"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vars"
)

// fileInfo exposes metadata from the os.Stat/os.Lstat functions.
//
// We ignore a few fs.FileMode bits, such as fs.ModeAppend, because they aren't
// found on platforms currently supported by Elvish; e.g., Plan9.
type fileInfo struct {
	// These fields are populated on every platform. However, they might only be
	// meaningful on some platforms. For example, on Windows the IsNamedPipe
	// member will always be false.
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
	MTime         time.Time
	// These fields are only populated on some platforms.
	ATime      time.Time
	BTime      time.Time
	CTime      time.Time
	Owner      string
	Group      string
	Uid        *big.Int
	Gid        *big.Int
	NumLinks   *big.Int
	Inode      *big.Int
	Device     *big.Int
	RawDevice  *big.Int
	BlockSize  *big.Int
	BlockCount *big.Int
}

func (fileInfo) IsStructMap() {}

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
		"stat":          stat,
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
