package eval

import (
	"strconv"
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
	for _, b := range builtinFns {
		ns[FnPrefix+b.Name] = NewRoVariable(b)
	}
	return ns
}
