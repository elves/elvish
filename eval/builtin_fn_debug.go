package eval

import (
	"runtime"

	"github.com/elves/elvish/util"
)

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
