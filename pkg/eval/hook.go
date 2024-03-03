package eval

import (
	"fmt"
	"os"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
)

// CallHook runs all the functions in the list "hook", with "args". If
// "cfg" is not specified, current eval's ports will be used.
func CallHook(ev *Evaler, cfg *EvalCfg, name string, hook vals.List, args ...any) {
	if hook.Len() == 0 {
		return
	}

	if cfg == nil {
		ports, cleanup := PortsFromStdFiles(ev.ValuePrefix())
		cfg = &EvalCfg{Ports: ports[:]}
		defer cleanup()
	}

	callCfg := CallCfg{Args: args, From: "[hook " + name + "]"}

	i := -1
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		fn, ok := it.Elem().(Callable)
		if !ok {
			diag.ShowError(os.Stderr, fmt.Errorf("hook %s[%d] must be callable", name, i))
			continue
		}

		err := ev.Call(fn, callCfg, *cfg)
		if err != nil {
			diag.ShowError(os.Stderr, err)
		}
	}
}
