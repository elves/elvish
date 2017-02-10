package edit

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

// Interface between the editor and elvish script. Implements the le: module.

var errNotNav = errors.New("not in navigation mode")

// makeModule builds a module from an Editor.
func makeModule(ed *Editor) eval.Namespace {
	ns := eval.Namespace{}
	// Populate builtins.
	for _, b := range builtins {
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
	ns["completer"] = eval.NewRoVariable(CompleterTable(argCompleter))
	ns[eval.FnPrefix+"complete-getopt"] = eval.NewRoVariable(
		// XXX Repr is "&le:complete-getopt" instead of "le:&complete-getopt"
		&eval.BuiltinFn{"le:complete-getopt", eval.WrapFn(complGetopt)})
	ns[eval.FnPrefix+"complete-files"] = eval.NewRoVariable(
		&eval.BuiltinFn{"le:complete-filename", eval.WrapFn(complFilenameFn)})

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

	ns["before-readline"] = ed.beforeReadLine
	ns["after-readline"] = ed.afterReadLine

	ns[eval.FnPrefix+"styled"] = eval.NewRoVariable(&eval.BuiltinFn{"le:styled", eval.WrapFn(styledBuiltin)})

	return ns
}

func throw(e error) {
	util.Throw(e)
}

func maybeThrow(e error) {
	if e != nil {
		util.Throw(e)
	}
}

func throwf(format string, args ...interface{}) {
	util.Throw(fmt.Errorf(format, args...))
}

// StringExposer implements eval.Variable and exposes a string to elvishscript.
type StringExposer struct {
	valuePtr *string
}

var errMustBeString = errors.New("must be string")

func (se StringExposer) Set(v eval.Value) {
	if s, ok := v.(eval.String); ok {
		*se.valuePtr = string(s)
	} else {
		throw(errMustBeString)
	}
}

func (se StringExposer) Get() eval.Value {
	return eval.String(*se.valuePtr)
}
