package eval

import (
	"strconv"
	"syscall"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/eval/vars"
)

var builtinNs = BuildNsNamed("").AddVars(map[string]vars.Var{
	"_":              vars.NewBlackhole(),
	"pid":            vars.NewReadOnly(strconv.Itoa(syscall.Getpid())),
	"ok":             vars.NewReadOnly(OK),
	"nil":            vars.NewReadOnly(nil),
	"true":           vars.NewReadOnly(true),
	"false":          vars.NewReadOnly(false),
	"buildinfo":      vars.NewReadOnly(buildinfo.Value),
	"version":        vars.NewReadOnly(buildinfo.Value.Version),
	"paths":          vars.NewEnvListVar("PATH"),
	"nop" + FnSuffix: vars.NewReadOnly(nopGoFn),
})

func addBuiltinFns(fns map[string]any) {
	builtinNs.AddGoFns(fns)
}
