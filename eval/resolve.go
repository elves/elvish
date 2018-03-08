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
	case "builtin":
		for name := range ev.Builtin {
			f(name)
		}
	case "":
		for name := range ev.Global {
			f(name)
		}
		for name := range ev.Builtin {
			f(name)
		}
	case "e":
		EachExternal(func(cmd string) {
			f(cmd + FnSuffix)
		})
	case "E":
		for _, s := range os.Environ() {
			if i := strings.IndexByte(s, '='); i > 0 {
				f(s[:i])
			}
		}
	default:
		segs := splitQName(ns + NsSuffix)
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
	f("builtin")
	f("e")
	f("E")

	for name := range ev.Global {
		if strings.HasSuffix(name, NsSuffix) {
			f(name[:len(name)-len(NsSuffix)])
		}
	}
	for name := range ev.Builtin {
		if strings.HasSuffix(name, NsSuffix) {
			f(name[:len(name)-len(NsSuffix)])
		}
	}
}

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (fm *Frame) ResolveVar(n, name string) vars.Var {
	if n == "" {
		return fm.resolveUnqualified(name)
	}

	// TODO: Let this function accept the fully qualified name.
	segs := splitQName(n + ":" + name)

	var ns Ns

	switch segs[0] {
	case "e:":
		if len(segs) == 2 && strings.HasSuffix(segs[1], FnSuffix) {
			return vars.NewRo(ExternalCmd{Name: segs[1][:len(segs[1])-len(FnSuffix)]})
		}
		return nil
	case "E:":
		if len(segs) == 2 {
			return vars.NewEnv(segs[1])
		}
		return nil
	case "local:":
		ns = fm.local
	case "up:":
		ns = fm.up
	case "builtin:":
		ns = fm.Builtin
	default:
		v := fm.resolveUnqualified(segs[0])
		if v == nil {
			return nil
		}
		ns = v.Get().(Ns)
	}

	for _, seg := range segs[1 : len(segs)-1] {
		v := ns[seg]
		if v == nil {
			return nil
		}
		ns = v.Get().(Ns)
	}
	return ns[segs[len(segs)-1]]
}

func splitQName(qname string) []string {
	i := 0
	var segs []string
	for i < len(qname) {
		j := strings.IndexByte(qname[i:], ':')
		if j == -1 {
			segs = append(segs, qname[i:])
			break
		}
		segs = append(segs, qname[i:i+j+1])
		i += j + 1
	}
	return segs
}

func (fm *Frame) resolveUnqualified(name string) vars.Var {
	if v, ok := fm.local[name]; ok {
		return v
	}
	if v, ok := fm.up[name]; ok {
		return v
	}
	return fm.Builtin[name]
}
