package edit

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/hashmap"
)

// This file contains several "registries", data structure that are written
// during program initialization and later used when initializing the Editor.
//
// The purpose of these registries is to decentralize the definition for
// builtins, default bindings, and variables (e.g. $edit:prompt). For instance,
// the definition for $edit:prompt can live in prompt.go instead of api.go.

var variableRegistry = map[string]func() eval.Variable{}

// RegisterVariable registers a variable: its name and a func to derive a
// Variable instance. It is later to be used during Editor initialization to
// populate Editor.variables as well as the edit: namespace.
func RegisterVariable(name string, maker func() eval.Variable) struct{} {
	variableRegistry[name] = maker
	return struct{}{}
}

func makeVariables() map[string]eval.Variable {
	m := make(map[string]eval.Variable, len(variableRegistry))
	for name, maker := range variableRegistry {
		m[name] = maker()
	}
	return m
}

var builtinMaps = map[string]map[string]*BuiltinFn{}

// registerBuiltins registers builtins under a subnamespace of edit:, to be used
// during the initialization of the Editor. It should be called for global
// variable initializations to make sure every subnamespace is registered before
// makeBindings is ever called.
func registerBuiltins(module string, impls map[string]func(*Editor)) struct{} {
	if _, ok := builtinMaps[module]; !ok {
		builtinMaps[module] = make(map[string]*BuiltinFn)
	}
	for name, impl := range impls {
		ns := "edit"
		if module != "" {
			ns += ":" + module
		}
		fullName := ns + ":" + name + eval.FnSuffix
		builtinMaps[module][name] = &BuiltinFn{fullName, impl}
	}
	return struct{}{}
}

func makeNsFromBuiltins(builtins map[string]*BuiltinFn) eval.Ns {
	ns := make(eval.Ns)
	for name, builtin := range builtins {
		ns[name+eval.FnSuffix] = eval.NewPtrVariable(builtin)
	}
	return ns
}

var keyBindings = map[string]map[ui.Key]eval.Fn{}

// registerBindings registers default bindings for a mode to initialize the
// global keyBindings map. Builtin names are resolved in the defaultMod
// subnamespace using information from builtinMaps. It should be called in init
// functions.
func registerBindings(
	mt string, defaultMod string, bindingData map[ui.Key]string) struct{} {

	if _, ok := keyBindings[mt]; !ok {
		keyBindings[mt] = map[ui.Key]eval.Fn{}
	}
	for key, fullName := range bindingData {
		// break fullName into mod and name.
		var mod, name string
		nameParts := strings.SplitN(fullName, ":", 2)
		if len(nameParts) == 2 {
			mod, name = nameParts[0], nameParts[1]
		} else {
			mod, name = defaultMod, nameParts[0]
		}
		if m, ok := builtinMaps[mod]; ok {
			if builtin, ok := m[name]; ok {
				keyBindings[mt][key] = builtin
			} else {
				fmt.Fprintln(os.Stderr, "Internal warning: no such builtin", name, "in mod", mod)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Internal warning: no such mod:", mod)
		}
	}
	return struct{}{}
}

func makeBindings() map[string]eval.Variable {
	bindings := make(map[string]eval.Variable)
	for mode, binding := range keyBindings {
		bindingValue := hashmap.Empty
		for key, fn := range binding {
			bindingValue = bindingValue.Assoc(key, fn)
		}
		bindings[mode] = eval.NewPtrVariableWithValidator(
			BindingTable{eval.NewMap(bindingValue)}, shouldBeBindingTable)
	}
	return bindings
}

var errShouldBeBindingTable = errors.New("should be binding table")

func shouldBeBindingTable(v eval.Value) error {
	if _, ok := v.(BindingTable); !ok {
		return errShouldBeBindingTable
	}
	return nil
}
