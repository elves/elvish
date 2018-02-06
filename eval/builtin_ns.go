package eval

import (
	"strconv"
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
	AddReflectBuiltinFns(ns, "", builtinFns)
	return ns
}

// AddReflectBuiltinFns adds reflect builtin functions to a namespace.
func AddReflectBuiltinFns(ns Ns, nsName string, fns map[string]interface{}) {
	for name, impl := range fns {
		qname := name
		if nsName != "" {
			qname = nsName + ":" + name
		}
		ns.SetFn(name, NewBuiltinFn(qname, impl))
	}
}
