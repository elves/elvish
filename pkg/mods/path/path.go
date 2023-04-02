// Package path provides functions for manipulating filesystem path names.
package path

import (
	_ "embed"
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
type fileInfo struct {
	// These fields are populated on every platform.
	Path         string
	AbsPath      string
	IsDir        bool
	Size         *big.Int
	Mode         *big.Int
	SymbolicMode string
	MTime        time.Time
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
