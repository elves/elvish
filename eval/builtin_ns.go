package eval

import (
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/eval/vartypes"
)

func makeBuiltinNs() Ns {
	ns := Ns{
		"_":     vartypes.NewBlackhole(),
		"pid":   vartypes.NewRo(strconv.Itoa(syscall.Getpid())),
		"ok":    vartypes.NewRo(OK),
		"true":  vartypes.NewRo(true),
		"false": vartypes.NewRo(false),
		"paths": &EnvList{envName: "PATH"},
		"pwd":   PwdVariable{},
	}
	AddBuiltinFns(ns, builtinFns...)
	return ns
}

// AddBuiltinFns adds builtin functions to a namespace.
func AddBuiltinFns(ns Ns, fns ...*BuiltinFn) {
	for _, b := range fns {
		name := b.Name
		if i := strings.IndexRune(b.Name, ':'); i != -1 {
			name = b.Name[i+1:]
		}
		ns[name+FnSuffix] = vartypes.NewRo(b)
	}
}
