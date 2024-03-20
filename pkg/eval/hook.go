package eval

import (
	"fmt"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
)

// CallHook runs all the functions in the list "hook", with "args".
//
// If "evalCfg" is not specified, the standard files will be used for IO.
//
// TODO: Eventually all callers should supply evalCfg. In general it's not
// correct to use standard files:
//
// - Chdir hooks should use the frame from which the chdir is triggered.
// - Editor lifecycle hooks should use the editor's TTY.
func CallHook(ev *Evaler, evalCfg *EvalCfg, name string, hook vals.List, args ...any) {
	if hook.Len() == 0 {
		return
	}

	if evalCfg == nil {
		ports, cleanup := PortsFromStdFiles(ev.ValuePrefix())
		defer cleanup()
		evalCfg = &EvalCfg{Ports: ports}
	}

	callCfg := CallCfg{Args: args, From: "[hook " + name + "]"}

	i := -1
	stderr := evalCfg.Ports[2].File
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		fn, ok := it.Elem().(Callable)
		if !ok {
			diag.ShowError(stderr, fmt.Errorf("hook %s[%d] must be callable", name, i))
			continue
		}

		err := ev.Call(fn, callCfg, *evalCfg)
		if err != nil {
			diag.ShowError(stderr, err)
		}
	}
}
