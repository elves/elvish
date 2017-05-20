package edit

import (
	"errors"

	"github.com/elves/elvish/eval"
)

// API for accessing the line editor from elvishscript, implemented as multiple
// modules.

var (
	errNotNav       = errors.New("not in navigation mode")
	errMustBeString = errors.New("must be string")
)

// installModules installs le: and le:* modules.
func installModules(modules map[string]eval.Namespace, ed *Editor) {
	// Construct the le: module.
	ns := eval.Namespace{}
	// Populate builtins.
	for _, b := range builtinMaps[""] {
		ns[eval.FnPrefix+b.name] = eval.NewPtrVariable(b)
	}

	// Populate binding tables in the variable $binding.
	// TODO Make binding specific to the Editor.
	binding := &eval.Struct{
		[]string{"insert", "command", "completion", "navigation", "history"},
		[]eval.Variable{
			eval.NewRoVariable(BindingTable{keyBindings[modeInsert]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCommand]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeCompletion]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeNavigation]}),
			eval.NewRoVariable(BindingTable{keyBindings[modeHistory]}),
		},
	}
	ns["binding"] = eval.NewRoVariable(binding)

	ns["completer"] = argCompleter
	ns[eval.FnPrefix+"complete-getopt"] = eval.NewRoVariable(
		&eval.BuiltinFn{"le:&complete-getopt", complGetopt})
	for _, bac := range argCompletersData {
		ns[eval.FnPrefix+bac.name] = eval.NewRoVariable(bac)
	}

	ns["prompt"] = ed.prompt
	ns["rprompt"] = ed.rprompt
	ns["rprompt-persistent"] = ed.rpromptPersistent
	ns["history"] = eval.NewRoVariable(History{&ed.historyMutex, ed.store})

	ns["current-command"] = eval.MakeVariableFromCallback(
		func(v eval.Value) {
			if !ed.active {
				throw(ErrEditorInactive)
			}
			if s, ok := v.(eval.String); ok {
				ed.line = string(s)
				ed.dot = len(ed.line)
			} else {
				throw(errMustBeString)
			}
		},
		func() eval.Value { return eval.String(ed.line) },
	)
	ns["selected-file"] = eval.MakeRoVariableFromCallback(
		func() eval.Value {
			if !ed.active {
				throw(ErrEditorInactive)
			}
			if ed.mode.Mode() != modeNavigation {
				throw(errNotNav)
			}
			return eval.String(ed.navigation.current.selectedName())
		},
	)

	ns["abbr"] = eval.NewRoVariable(eval.MapStringString(ed.abbreviations))

	ns["loc-pinned"] = ed.locationPinned
	ns["loc-hidden"] = ed.locationHidden

	ns["before-readline"] = ed.beforeReadLine
	ns["after-readline"] = ed.afterReadLine

	ns[eval.FnPrefix+"styled"] = eval.NewRoVariable(&eval.BuiltinFn{"le:&styled", styledBuiltin})

	modules["le"] = ns

	// Install other modules.
	for module, builtins := range builtinMaps {
		if module != "" {
			modules["le:"+module] = makeNamespaceFromBuiltins(builtins)
		}
	}
}
