package complete

import (
	"os"
	"strings"

	"src.elv.sh/pkg/eval"
)

var environ = os.Environ

// Calls the passed function for each variable name in namespace ns that can be
// found from the top context.
func eachVariableInNs(ev *eval.Evaler, ns string, f func(s string)) {
	switch ns {
	case "", ":":
		ev.Global().IterateKeysString(f)
		ev.Builtin().IterateKeysString(f)
	case "e:":
		eachExternal(func(cmd string) {
			f(cmd + eval.FnSuffix)
		})
	case "E:":
		for _, s := range environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		segs := eval.SplitQNameSegs(ns)
		mod := ev.Global().IndexString(segs[0])
		if mod == nil {
			mod = ev.Builtin().IndexString(segs[0])
		}
		for _, seg := range segs[1:] {
			if mod == nil {
				return
			}
			mod = mod.Get().(*eval.Ns).IndexString(seg)
		}
		if mod != nil {
			mod.Get().(*eval.Ns).IterateKeysString(f)
		}
	}
}
