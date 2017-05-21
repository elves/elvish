package edit

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// This file contains two "registries", data structure that are written during
// program initialization by implementations of different modes and later used
// when initializing the Editor.

var (
	builtinMaps = map[string]map[string]*BuiltinFn{}
	keyBindings = map[ModeType]map[ui.Key]eval.CallableValue{}
)

// registerBuiltins makes a map of builtins from a map of implementations.
// It also records the made map in builtinMaps for use in makeBindings. It
// should be called for global variable initialization to make sure every map is
// registered before makeBindings is ever called.
func registerBuiltins(module string, impls map[string]func(*Editor)) struct{} {
	builtinMaps[module] = make(map[string]*BuiltinFn)
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

// registerBindings makes a map from Key to CallableValue from a map from Key to
// string, resolving the builtins using builtinMaps. For unqualified names, it
// assumes the passed module name.
func registerBindings(
	mt ModeType, defaultMod string,
	bindingData map[ui.Key]string) struct{} {

	bindings := map[ui.Key]eval.CallableValue{}
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
				bindings[key] = builtin
			} else {
				fmt.Fprintln(os.Stderr, "Internal warning: no such builtin", name, "in mod", mod)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Internal warning: no such mod:", mod)
		}
	}
	keyBindings[mt] = bindings
	return struct{}{}
}
