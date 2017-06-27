package eval

import (
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/daemon/api"
)

func makeBuiltinNamespace(daemon *api.Client) Namespace {
	ns := Namespace{
		"pid":   NewRoVariable(String(strconv.Itoa(syscall.Getpid()))),
		"ok":    NewRoVariable(OK),
		"true":  NewRoVariable(Bool(true)),
		"false": NewRoVariable(Bool(false)),
		"paths": &EnvPathList{envName: "PATH"},
		"pwd":   PwdVariable{daemon},
	}
	AddBuiltinFns(ns, builtinFns...)
	return ns
}

// AddBuiltinFns adds builtin functions to a namespace.
func AddBuiltinFns(ns Namespace, fns ...*BuiltinFn) {
	for _, b := range fns {
		name := b.Name
		if i := strings.IndexRune(b.Name, ':'); i != -1 {
			name = b.Name[i+1:]
		}
		ns[FnPrefix+name] = NewRoVariable(b)
	}
}
