package eval

import (
	"runtime"

	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/parse"
)

func init() {
	addBuiltinFns(map[string]interface{}{
		"src":    src,
		"-gc":    _gc,
		"-stack": _stack,
		"-log":   _log,
	})
}

//elvdoc:fn src
//
// ```elvish
// src
// ```
//
// Output a map-like value describing the current source being evaluated. The value
// contains the following fields:
//
// -   `name`, a unique name of the current source. If the source originates from a
//     file, it is the full path of the file.
//
// -   `code`, the full body of the current source.
//
// -   `is-file`, whether the source originates from a file.
//
// Examples:
//
// ```elvish-transcript
// ~> put (src)[name code is-file]
// ▶ '[tty]'
// ▶ 'put (src)[name code is-file]'
// ▶ $false
// ~> echo 'put (src)[name code is-file]' > show-src.elv
// ~> elvish show-src.elv
// ▶ /home/elf/show-src.elv
// ▶ "put (src)[name code is-file]\n"
// ▶ $true
//
// Note: this builtin always returns information of the source of the function
// calling `src`. Consider the following example:
//
// ```elvish-transcript
// ~> echo 'fn show { put (src)[name] }' > ~/.elvish/lib/src-fsutil.elv
// ~> use src-util
// ~> src-util:show
// ▶ /home/elf/.elvish/lib/src-fsutil.elv
// ```

func src(fm *Frame) parse.Source {
	return fm.srcMeta
}

//elvdoc:fn -gc
//
// ```elvish
// -gc
// ```
//
// Force the Go garbage collector to run.
//
// This is only useful for debug purposes.

func _gc() {
	runtime.GC()
}

//elvdoc:fn -stack
//
// ```elvish
// -stack
// ```
//
// Print a stack trace.
//
// This is only useful for debug purposes.

func _stack(fm *Frame) {
	out := fm.OutputFile()
	// TODO(xiaq): Dup with main.go.
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}

//elvdoc:fn -log
//
// ```elvish
// -log $filename
// ```
//
// Direct internal debug logs to the named file.
//
// This is only useful for debug purposes.

func _log(fname string) error {
	return logutil.SetOutputFile(fname)
}
