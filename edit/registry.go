package edit

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// This file contains several "registries", data structure that are written
// during program initialization and later used when initializing the Editor.

var builtinMaps = map[string]map[string]*BuiltinFn{}

// registerBuiltins registers builtins under a subnamespace of le:, to be used
// during the initialization of the Editor. It should be called for global
// variable initializations to make sure every subnamespace is registered before
// makeBindings is ever called.
func registerBuiltins(module string, impls map[string]func(*Editor)) struct{} {
	if _, ok := builtinMaps[module]; !ok {
		builtinMaps[module] = make(map[string]*BuiltinFn)
	}
	for name, impl := range impls {
		var fullName string
		if module == "" {
			fullName = "le:" + eval.FnPrefix + name
		} else {
			fullName = "le:" + module + ":" + eval.FnPrefix + name
		}
		builtinMaps[module][name] = &BuiltinFn{fullName, impl}
	}
	return struct{}{}
}

func makeNamespaceFromBuiltins(builtins map[string]*BuiltinFn) eval.Namespace {
	ns := eval.Namespace{}
	for name, builtin := range builtins {
		ns[eval.FnPrefix+name] = eval.NewPtrVariable(builtin)
	}
	return ns
}

var keyBindings = map[string]map[ui.Key]eval.CallableValue{}

// registerBindings registers default bindings for a mode to initialize the
// global keyBindings map. Builtin names are resolved in the defaultMod
// subnamespace using information from builtinMaps. It should be called in init
// functions.
func registerBindings(
	mt string, defaultMod string,
	bindingData map[ui.Key]string) struct{} {

	if _, ok := keyBindings[mt]; !ok {
		keyBindings[mt] = map[ui.Key]eval.CallableValue{}
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

var variableMakers = map[string]func() eval.Variable{}

// registerVariables registers a variable, its name and a func used to derive
// its value, later to be used during Editor initialization to populate
// Editor.variables as well as the le: namespace.
func registerVariable(name string, maker func() eval.Variable) struct{} {
	variableMakers[name] = maker
	return struct{}{}
}

func makeVariables() map[string]eval.Variable {
	m := make(map[string]eval.Variable, len(variableMakers))
	for name, maker := range variableMakers {
		m[name] = maker()
	}
	return m
}
