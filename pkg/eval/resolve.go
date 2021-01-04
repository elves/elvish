package eval

import (
	"os"
	"strings"

	"github.com/elves/elvish/pkg/fsutil"
)

// EachVariableInTop calls the passed function for each variable name in
// namespace ns that can be found from the top context.
func EachVariableInTop(builtin, global *Ns, ns string, f func(s string)) {
	switch ns {
	case "builtin:":
		for _, name := range builtin.names {
			f(name)
		}
	case "", ":":
		for _, name := range global.names {
			f(name)
		}
		for _, name := range builtin.names {
			f(name)
		}
	case "e:":
		fsutil.EachExternal(func(cmd string) {
			f(cmd + FnSuffix)
		})
	case "E:":
		for _, s := range os.Environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		segs := SplitQNameSegs(ns)
		mod := global.indexInner(segs[0])
		if mod == nil {
			mod = builtin.indexInner(segs[0])
		}
		for _, seg := range segs[1:] {
			if mod == nil {
				return
			}
			mod = mod.Get().(*Ns).indexInner(seg)
		}
		if mod != nil {
			for _, name := range mod.Get().(*Ns).names {
				f(name)
			}
		}
	}
}

// EachNsInTop calls the passed function for each namespace that can be used
// from the top context.
func EachNsInTop(builtin, global *Ns, f func(s string)) {
	f("builtin:")
	f("e:")
	f("E:")

	for _, name := range global.names {
		if strings.HasSuffix(name, NsSuffix) {
			f(name)
		}
	}
	for _, name := range builtin.names {
		if strings.HasSuffix(name, NsSuffix) {
			f(name)
		}
	}
}
