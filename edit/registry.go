package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

// This file contains several "registries", data structure that are written
// during program initialization and later used when initializing the Editor.
//
// The purpose of these registries is to decentralize the definition for
// builtins, default bindings, and variables (e.g. $edit:prompt). For instance,
// the definition for $edit:prompt can live in prompt.go instead of api.go.

var editorInitFuncs []func(*Editor)

func atEditorInit(f func(*Editor)) {
	editorInitFuncs = append(editorInitFuncs, f)
}

var builtinMaps = map[string]map[string]func(*Editor){}

// registerBuiltins registers builtins under a subnamespace of edit:, to be used
// during the initialization of the Editor. It should be called for global
// variable initializations to make sure every subnamespace is registered before
// makeBindings is ever called.
func registerBuiltins(module string, impls map[string]func(*Editor)) struct{} {
	builtinMaps[module] = impls
	return struct{}{}
}

func makeNsFromBuiltins(nsName string, builtins map[string]func(*Editor)) eval.Ns {
	ns := eval.NewNs()
	for name, impl := range builtins {
		ns.AddFn(name, &BuiltinFn{"edit:" + nsName + name, impl})
	}
	return ns
}

func makeBindings() map[string]vartypes.Variable {
	bindings := map[string]vartypes.Variable{}
	// XXX This abuses the builtin registry to get a list of mode names
	for mode := range builtinMaps {
		table := BindingTable{types.EmptyMap}
		bindings[mode] = eval.NewVariableFromPtr(&table)
	}
	return bindings
}
