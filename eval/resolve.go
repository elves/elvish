package eval

import (
	"os"
	"strings"

	"github.com/elves/elvish/eval/vars"
)

// Resolution and iteration of variables and namespaces.

// EachVariableInTop calls the passed function for each variable name in
// namespace ns that can be found from the top context.
func (ev *evalerScopes) EachVariableInTop(ns string, f func(s string)) {
	switch ns {
	case "builtin:":
		for name := range ev.Builtin {
			f(name)
		}
	case "", ":":
		for name := range ev.Global {
			f(name)
		}
		for name := range ev.Builtin {
			f(name)
		}
	case "e:":
		EachExternal(func(cmd string) {
			f(cmd + FnSuffix)
		})
	case "E:":
		for _, s := range os.Environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		segs := SplitQNameNsSegs(ns)
		mod := ev.Global[segs[0]]
		if mod == nil {
			mod = ev.Builtin[segs[0]]
		}
		for _, seg := range segs[1:] {
			if mod == nil {
				return
			}
			mod = mod.Get().(Ns)[seg]
		}
		if mod != nil {
			for name := range mod.Get().(Ns) {
				f(name)
			}
		}
	}
}

// EachNsInTop calls the passed function for each namespace that can be used
// from the top context.
func (ev *evalerScopes) EachNsInTop(f func(s string)) {
	f("builtin:")
	f("e:")
	f("E:")

	for name := range ev.Global {
		if strings.HasSuffix(name, NsSuffix) {
			f(name)
		}
	}
	for name := range ev.Builtin {
		if strings.HasSuffix(name, NsSuffix) {
			f(name)
		}
	}
}

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (fm *Frame) ResolveVar(qname string) vars.Var {
	ns, name := SplitQNameNsFirst(qname)

	switch ns {
	case "E:":
		return vars.FromEnv(name)
	case "e:":
		if strings.HasSuffix(name, FnSuffix) {
			return vars.NewReadOnly(ExternalCmd{name[:len(name)-len(FnSuffix)]})
		}
		return nil
	case "local:":
		return resolveNested(fm.local, name)
	case "up:":
		return resolveNested(fm.up, name)
	case "builtin:":
		return resolveNested(fm.Builtin, name)
	case "", ":":
		return fm.resolveNonPseudo(name)
	default:
		return fm.resolveNonPseudo(qname)
	}
}

func (fm *Frame) resolveNonPseudo(name string) vars.Var {
	if v := resolveNested(fm.local, name); v != nil {
		return v
	}
	if v := resolveNested(fm.up, name); v != nil {
		return v
	}
	return resolveNested(fm.Builtin, name)
}

func resolveNested(ns Ns, name string) vars.Var {
	if name == "" {
		return nil
	}
	segs := SplitQNameNsSegs(name)
	for _, seg := range segs[:len(segs)-1] {
		variable := ns[seg]
		if variable == nil {
			return nil
		}
		nestedNs, ok := variable.Get().(Ns)
		if !ok {
			return nil
		}
		ns = nestedNs
	}
	return ns[segs[len(segs)-1]]
}
