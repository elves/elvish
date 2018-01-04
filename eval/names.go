package eval

import (
	"errors"
	"os"
	"strings"

	"github.com/elves/elvish/eval/vartypes"
)

// Resolution and iteration of variables and namespaces.

// ErrStoreUnconnected is thrown by ResolveVar when a shared: variable needs to
// be resolved but the store is not connected.
var ErrStoreUnconnected = errors.New("store unconnected")

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
	case "shared":
		// TODO Add daemon RPC for enumerating shared variables.
	default:
		mod := ev.Global[ns+NsSuffix]
		if mod == nil {
			mod = ev.Builtin[ns+NsSuffix]
		}
		if mod != nil {
			for name := range mod.Get().(Ns) {
				f(name)
			}
		}
	}
}

// EachModInTop calls the passed function for each module that can be found from
// the top context.
func (ev *evalerScopes) EachModInTop(f func(s string)) {
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

// EachNsInTop calls the passed function for each namespace that can be used
// from the top context.
func (ev *evalerScopes) EachNsInTop(f func(s string)) {
	f("builtin")
	f("e")
	f("E")
	f("shared")
	ev.EachModInTop(f)
}

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (ec *Frame) ResolveVar(ns, name string) vartypes.Variable {
	switch ns {
	case "local":
		return ec.local[name]
	case "up":
		return ec.up[name]
	case "builtin":
		return ec.Builtin[name]
	case "":
		if v := ec.local[name]; v != nil {
			return v
		}
		if v, ok := ec.up[name]; ok {
			return v
		}
		return ec.Builtin[name]
	case "e":
		if strings.HasSuffix(name, FnSuffix) {
			return vartypes.NewRo(ExternalCmd{name[:len(name)-len(FnSuffix)]})
		}
	case "E":
		return vartypes.NewEnv(name)
	case "shared":
		if ec.DaemonClient == nil {
			throw(ErrStoreUnconnected)
		}
		return sharedVariable{ec.DaemonClient, name}
	default:
		ns := ec.ResolveMod(ns)
		if ns != nil {
			return ns[name]
		}
	}
	return nil
}

func (ec *Frame) ResolveMod(name string) Ns {
	ns := ec.ResolveVar("", name+NsSuffix)
	if ns == nil {
		return nil
	}
	return ns.Get().(Ns)
}
