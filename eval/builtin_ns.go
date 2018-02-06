package eval

import (
	"strconv"
	"syscall"

	"github.com/elves/elvish/eval/vartypes"
)

var builtinNs = Ns{
	"_":     vartypes.NewBlackhole(),
	"pid":   vartypes.NewRo(strconv.Itoa(syscall.Getpid())),
	"ok":    vartypes.NewRo(OK),
	"true":  vartypes.NewRo(true),
	"false": vartypes.NewRo(false),
	"paths": &EnvList{envName: "PATH"},
	"pwd":   PwdVariable{},
}

func addBuiltinFns(fns map[string]interface{}) {
	builtinNs.AddBuiltinFns("", fns)
}
