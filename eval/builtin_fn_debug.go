package eval

import (
	"runtime"

	"github.com/elves/elvish/util"
)

func init() {
	addToBuiltinFns([]*BuiltinFn{
		// Debugging
		{"src", src},
		{"-gc", _gc},
		{"-stack", _stack},
		{"-log", _log},
	})
}

func src(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	ec.OutputChan() <- ec.srcMeta
}

func _gc(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	runtime.GC()
}

func _stack(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	out := ec.ports[1].File
	// XXX dup with main.go
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	out.Write(buf)
}

func _log(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var fnamev string
	ScanArgs(args, &fnamev)
	fname := fnamev
	TakeNoOpt(opts)

	maybeThrow(util.SetOutputFile(fname))
}
