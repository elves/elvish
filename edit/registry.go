package edit

import (
	. "github.com/elves/elvish/edit/edtypes"
	"github.com/elves/elvish/eval"
)

// This file contains utilities that facilitates modularization of the editor.

var editorInitFuncs []func(*editor, eval.Ns)

func atEditorInit(f func(*editor, eval.Ns)) {
	editorInitFuncs = append(editorInitFuncs, f)
}

func makeNsFromBuiltins(nsName string, builtins map[string]func(*editor)) eval.Ns {
	ns := eval.NewNs()
	for name, impl := range builtins {
		ns.AddFn(name, &BuiltinFn{"edit:" + nsName + name, impl})
	}
	return ns
}

func initModeAPI(n string, f map[string]func(*editor), p *BindingMap) eval.Ns {
	*p = EmptyBindingMap
	return makeNsFromBuiltins(n, f).Add("binding", eval.NewVariableFromPtr(p))
}
