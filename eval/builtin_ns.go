package eval

import (
	"strconv"
	"syscall"

	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/vector"
)

var builtinNs = Ns{
	"_":     vars.NewBlackhole(),
	"pid":   vars.NewReadOnly(strconv.Itoa(syscall.Getpid())),
	"ok":    vars.NewReadOnly(OK),
	"true":  vars.NewReadOnly(true),
	"false": vars.NewReadOnly(false),
	"paths": &EnvList{envName: "PATH"},
	"pwd":   PwdVariable{},
	"args":  vars.NewReadOnly(vector.Empty),
}

func addBuiltinFns(fns map[string]interface{}) {
	builtinNs.AddBuiltinFns("", fns)
}
