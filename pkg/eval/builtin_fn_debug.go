package eval

import (
	"runtime"

	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/parse"
)

func init() {
	addBuiltinFns(map[string]any{
		"src":    src,
		"-gc":    _gc,
		"-stack": _stack,
		"-log":   _log,
	})
}

func src(fm *Frame) parse.Source {
	return fm.src
}

func _gc() {
	runtime.GC()
}

func _stack(fm *Frame) error {
	// TODO(xiaq): Dup with main.go.
	buf := make([]byte, 1024)
	for runtime.Stack(buf, true) == cap(buf) {
		buf = make([]byte, cap(buf)*2)
	}
	_, err := fm.ByteOutput().Write(buf)
	return err
}

func _log(fname string) error {
	return logutil.SetOutputFile(fname)
}
