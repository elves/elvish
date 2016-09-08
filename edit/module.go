package edit

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

// Interface between the editor and elvish script. Implements the le: module.

// makeModule builds a module from an Editor.
func makeModule(ed *Editor) eval.Namespace {
	ns := eval.Namespace{}
	// Populate builtins.
	for _, b := range builtins {
		ns[eval.FnPrefix+b.name] = eval.NewPtrVariable(&BuiltinAsFnValue{b, ed})
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

	ns["prompt"] = ed.ps1
	ns["rprompt"] = ed.rps1
	ns["rprompt-persistent"] = BoolExposer{&ed.rpromptPersistent}
	ns["current-command"] = LineExposer{ed}
	ns["history"] = eval.NewRoVariable(History{&ed.historyMutex, ed.store})

	ns["abbr"] = eval.NewRoVariable(eval.MapStringString(ed.abbreviations))

	ns["before-readline"] = ed.beforeReadLine

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

// BoolExposer implements eval.Variable and exposes a bool to elvishscript.
type BoolExposer struct {
	valuePtr *bool
}

var errMustBeBool = errors.New("must be bool")

func (be BoolExposer) Set(v eval.Value) {
	if b, ok := v.(eval.Bool); ok {
		*be.valuePtr = bool(b)
	} else {
		throw(errMustBeBool)
	}
}

func (be BoolExposer) Get() eval.Value {
	return eval.Bool(*be.valuePtr)
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

// LineExposer exposes ed.line to elvishscript.

type LineExposer struct {
	ed *Editor
}

func (l LineExposer) Set(v eval.Value) {
	if s, ok := v.(eval.String); ok {
		l.ed.line = string(s)
		l.ed.dot = len(l.ed.line)
	} else {
		throw(errMustBeString)
	}
}

func (l LineExposer) Get() eval.Value {
	return eval.String(l.ed.line)
}
