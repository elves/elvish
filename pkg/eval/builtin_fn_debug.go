package eval

import (
	"runtime"

	"github.com/elves/elvish/pkg/util"
)

//elvdoc:fn src
//
// ```elvish
// src
// ```
//
// Output a map-like value describing the current source being evaluated. The value
// contains the following fields:
//
// -   `type`, which can be one of `interactive`, `script` or `module`;
//
// -   `name`, which is set to the name under which a script is executed or a
// module is imported. It is an empty string when `type` = `interactive`;
//
// -   `path`, which is the path to the current source. It is an empty string when
// `type` = `interactive`;
//
// -   `code`, which is the full body of the current source.
//
// Examples:
//
// ```elvish-transcript
// ~> put (src)[type name path code]
// ▶ interactive
// ▶ ''
// ▶ ''
// ▶ 'put (src)[type name path code]'
// ~> echo 'put (src)[type name path code]' > foo.elv
// ~> elvish foo.elv
// ▶ script
// ▶ foo.elv
// ▶ /home/xiaq/foo.elv
// ▶ "put (src)[type name path code]\n"
// ~> echo 'put (src)[type name path code]' > ~/.elvish/lib/m.elv
// ~> use m
// ▶ module
// ▶ m
// ▶ /home/xiaq/.elvish/lib/m.elv
// ▶ "put (src)[type name path code]\n"
// ```
//
// Note: this builtin always returns information of the source of the **calling
// function**. Example:
//
// ```elvish-transcript
// ~> echo 'fn f { put (src)[type name path code] }' > ~/.elvish/lib/n.elv
// ~> use n
// ~> n:f
// ▶ module
// ▶ n
// ▶ /home/xiaq/.elvish/lib/n.elv
// ▶ "fn f { put (src)[type name path code] }\n"
// ```

//elvdoc:fn -gc
//
// ```elvish
// -gc
// ```
//
// Force the Go garbage collector to run.
//
// This is only useful for debug purposes.

//elvdoc:fn -stack
//
// ```elvish
// -stack
// ```
//
// Print a stack trace.
//
// This is only useful for debug purposes.

//elvdoc:fn -log
//
// ```elvish
// -log $filename
// ```
//
// Direct internal debug logs to the named file.
//
// This is only useful for debug purposes.

func init() {
	addBuiltinFns(map[string]interface{}{
		"src":    src,
		"-gc":    _gc,
		"-stack": _stack,
		"-log":   _log,
	})
}

func src(fm *Frame) *Source {
	return fm.srcMeta
}

func _gc() {
	runtime.GC()
}

func _stack(fm *Frame) {
	out := fm.ports[1].File
	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}

func _log(fname string) error {
	return util.SetOutputFile(fname)
}
