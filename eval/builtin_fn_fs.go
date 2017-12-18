package eval

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

// Filesystem.

var ErrStoreNotConnected = errors.New("store not connected")

func init() {
	addToBuiltinFns([]*BuiltinFn{
		// Directory
		{"cd", cd},
		{"dir-history", dirs},

		// Path
		{"path-abs", WrapStringToStringError(filepath.Abs)},
		{"path-base", WrapStringToString(filepath.Base)},
		{"path-clean", WrapStringToString(filepath.Clean)},
		{"path-dir", WrapStringToString(filepath.Dir)},
		{"path-ext", WrapStringToString(filepath.Ext)},

		{"eval-symlinks", WrapStringToStringError(filepath.EvalSymlinks)},
		{"tilde-abbr", tildeAbbr},

		// File types
		{"-is-dir", isDir},
	})
}

func WrapStringToString(f func(string) string) BuiltinFnImpl {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		ec.ports[1].Chan <- String(f(s))
	}
}

func WrapStringToStringError(f func(string) (string, error)) BuiltinFnImpl {
	return func(ec *EvalCtx, args []Value, opts map[string]Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		result, err := f(s)
		maybeThrow(err)
		ec.ports[1].Chan <- String(result)
	}
}

var errMustBeOneString = errors.New("must be one string argument")

func mustGetOneString(args []Value) string {
	if len(args) != 1 {
		throw(errMustBeOneString)
	}
	s, ok := args[0].(String)
	if !ok {
		throw(errMustBeOneString)
	}
	return string(s)
}

func cd(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoOpt(opts)

	var dir string
	if len(args) == 0 {
		dir = mustGetHome("")
	} else if len(args) == 1 {
		dir = ToString(args[0])
	} else {
		throw(ErrArgs)
	}

	cdInner(dir, ec)
}

func cdInner(dir string, ec *EvalCtx) {
	maybeThrow(Chdir(dir, ec.Daemon))
}

var dirDescriptor = NewStructDescriptor("path", "score")

func dirs(ec *EvalCtx, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	if ec.Daemon == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.Daemon.Dirs(storedefs.NoBlacklist)
	if err != nil {
		throw(errors.New("store error: " + err.Error()))
	}
	out := ec.ports[1].Chan
	for _, dir := range dirs {
		out <- &Struct{dirDescriptor, []Value{
			String(dir.Path),
			floatToString(dir.Score),
		}}
	}
}

func tildeAbbr(ec *EvalCtx, args []Value, opts map[string]Value) {
	var pathv String
	ScanArgs(args, &pathv)
	path := string(pathv)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- String(util.TildeAbbr(path))
}

func isDir(ec *EvalCtx, args []Value, opts map[string]Value) {
	var pathv String
	ScanArgs(args, &pathv)
	path := string(pathv)
	TakeNoOpt(opts)

	ec.OutputChan() <- Bool(isDirInner(path))
}

func isDirInner(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsDir()
}
