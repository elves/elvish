package eval

import (
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

func makeBuiltinNs() Ns {
	ns := Ns{
		"_":     vartypes.NewBlackhole(),
		"pid":   vartypes.NewRoVariable(types.String(strconv.Itoa(syscall.Getpid()))),
		"ok":    vartypes.NewRoVariable(OK),
		"true":  vartypes.NewRoVariable(types.Bool(true)),
		"false": vartypes.NewRoVariable(types.Bool(false)),
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
		ns[name+FnSuffix] = vartypes.NewRoVariable(b)
	}
}
