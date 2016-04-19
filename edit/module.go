package edit

import (
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

	ns["prompt"] = PromptVariable{&ed.ps1}
	ns["rprompt"] = PromptVariable{&ed.rps1}

	ns["abbr"] = eval.NewRoVariable(eval.MapStringString(ed.abbreviations))

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
