package eval

import (
	"strconv"
	"syscall"

	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/vector"
)

var builtinNs = Ns{
	"_":     vars.NewBlackhole(),
	"pid":   vars.NewRo(strconv.Itoa(syscall.Getpid())),
	"ok":    vars.NewRo(OK),
	"true":  vars.NewRo(true),
	"false": vars.NewRo(false),
	"paths": &EnvList{envName: "PATH"},
	"pwd":   PwdVariable{},
	"args":  vars.NewRo(vector.Empty),
}

func addBuiltinFns(fns map[string]interface{}) {
	builtinNs.AddBuiltinFns("", fns)
}
