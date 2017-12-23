package eval

import (
	"strconv"
	"strings"
	"syscall"
)

func makeBuiltinNs() Ns {
	ns := Ns{
		"_":     BlackholeVariable{},
		"pid":   NewRoVariable(String(strconv.Itoa(syscall.Getpid()))),
		"ok":    NewRoVariable(OK),
		"true":  NewRoVariable(Bool(true)),
		"false": NewRoVariable(Bool(false)),
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
		ns[name+FnSuffix] = NewRoVariable(b)
	}
}
