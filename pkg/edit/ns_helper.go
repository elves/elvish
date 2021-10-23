package edit

import (
	"os"
	"strings"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/fsutil"
)

// Calls the passed function for each variable name in namespace ns that can be
// found from the top context.
func eachVariableInTop(builtin, global *eval.Ns, ns string, f func(s string)) {
	switch ns {
	case "", ":":
		global.IterateKeysString(f)
		builtin.IterateKeysString(f)
	case "e:":
		fsutil.EachExternal(func(cmd string) {
			f(cmd + eval.FnSuffix)
		})
	case "E:":
		for _, s := range os.Environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		segs := eval.SplitQNameSegs(ns)
		mod := global.IndexString(segs[0])
		if mod == nil {
			mod = builtin.IndexString(segs[0])
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

// Calls the passed function for each namespace that can be used from the top
// context.
func eachNsInTop(builtin, global *eval.Ns, f func(s string)) {
	f("e:")
	f("E:")

	global.IterateKeysString(func(name string) {
		if strings.HasSuffix(name, eval.NsSuffix) {
			f(name)
		}
	})

	builtin.IterateKeysString(func(name string) {
		if strings.HasSuffix(name, eval.NsSuffix) {
			f(name)
		}
	})
}
