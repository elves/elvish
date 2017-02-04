package eval

import (
	"strconv"
	"syscall"
)

var builtinNamespace = Namespace{
	"pid":   NewRoVariable(String(strconv.Itoa(syscall.Getpid()))),
	"ok":    NewRoVariable(OK),
	"true":  NewRoVariable(Bool(true)),
	"false": NewRoVariable(Bool(false)),
	"paths": &EnvPathList{envName: "PATH"},
	"pwd":   PwdVariable{},
}
