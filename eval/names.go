package eval

import (
	"errors"
	"os"
	"strings"
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
		for name := range ev.Builtin.Names {
			f(name)
		}
	case "":
		for name := range ev.Global.Names {
			f(name)
		}
		for name := range ev.Builtin.Names {
			f(name)
		}
	case "e":
		EachExternal(func(cmd string) {
			f(FnPrefix + cmd)
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
		mod := ev.Global.Uses[ns]
		if mod == nil {
			mod = ev.Builtin.Uses[ns]
		}
		for name := range mod {
			f(name)
		}
	}
}

// EachModInTop calls the passed function for each module that can be found from
// the top context.
func (ev *evalerScopes) EachModInTop(f func(s string)) {
	for name := range ev.Global.Uses {
		f(name)
	}
	for name := range ev.Builtin.Uses {
		f(name)
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
func (ec *EvalCtx) ResolveVar(ns, name string) Variable {
	switch ns {
	case "local":
		return ec.local.Names[name]
	case "up":
		return ec.up.Names[name]
	case "builtin":
		return ec.Builtin.Names[name]
	case "":
		if v := ec.local.Names[name]; v != nil {
			return v
		}
		if v, ok := ec.up.Names[name]; ok {
			return v
		}
		return ec.Builtin.Names[name]
	case "e":
		if strings.HasPrefix(name, FnPrefix) {
			return NewRoVariable(ExternalCmd{name[len(FnPrefix):]})
		}
	case "E":
		return envVariable{name}
	case "shared":
		if ec.Daemon == nil {
			throw(ErrStoreUnconnected)
		}
		return sharedVariable{ec.Daemon, name}
	default:
		ns := ec.ResolveMod(ns)
		if ns != nil {
			return ns[name]
		}
	}
	return nil
}

func (ec *EvalCtx) ResolveMod(name string) Namespace {
	if ns, ok := ec.local.Uses[name]; ok {
		return ns
	}
	if ns, ok := ec.up.Uses[name]; ok {
		return ns
	}
	if ns, ok := ec.Builtin.Uses[name]; ok {
		return ns
	}
	return nil
}
