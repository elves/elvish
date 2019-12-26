// Package edit implements the line editor for Elvish.
//
// The line editor is based on the cli package, which implements a general,
// Elvish-agnostic line editor, and multiple "addon" packages. This package
// glues them together and provides Elvish bindings for them.
package edit

import (
	"os"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/store"
)

// Editor is the interface line editor for Elvish.
type Editor struct {
	app cli.App
	ns  eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(tty cli.TTY, ev *eval.Evaler, st store.Service) *Editor {
	ns := eval.NewNs()
	appSpec := cli.AppSpec{TTY: tty}

	fuser, err := histutil.NewFuser(st)
	if err != nil {
		// TODO(xiaq): Report the error.
	}

	if fuser != nil {
		appSpec.AfterReadline = []func(string){func(code string) {
			if code != "" {
				fuser.AddCmd(code)
			}
			// TODO(xiaq): Handle the error.
		}}
	}

	// Make a variable for the app first. This is to work around the
	// bootstrapping of initPrompts, which expects a notifier.
	var app cli.App
	initHighlighter(&appSpec, ev)
	initConfigAPI(&appSpec, ev, ns)
	initInsertAPI(&appSpec, appNotifier{&app}, ev, ns)
	initPrompts(&appSpec, appNotifier{&app}, ev, ns)
	app = cli.NewApp(appSpec)

	initCommandAPI(app, ev, ns)
	initListings(app, ev, ns, st, fuser)
	initNavigation(app, ev, ns)
	initCompletion(app, ev, ns)
	initHistWalk(app, ev, ns, fuser)
	initInstant(app, ev, ns)
	initMinibuf(app, ev, ns)

	initBufferBuiltins(app, ns)
	initTTYBuiltins(app, tty, ns)
	initMiscBuiltins(app, ns)
	initStateAPI(app, ns)
	initStoreAPI(app, ns, fuser)
	evalDefaultBinding(ev, ns)

	return &Editor{app, ns}
}

func evalDefaultBinding(ev *eval.Evaler, ns eval.Ns) {
	// TODO(xiaq): The evaler API should accodomate the use case of evaluating a
	// piece of code in an alternative global namespace.

	n, err := parse.AsChunk("[default bindings]", defaultBindingsElv)
	if err != nil {
		panic(err)
	}
	src := eval.NewScriptSource(
		"[default bindings]", "[default bindings]", defaultBindingsElv)
	op, err := ev.CompileWithGlobal(n, src, ns)
	if err != nil {
		panic(err)
	}
	// TODO(xiaq): Use stdPorts when it is possible to do so.
	fm := eval.NewTopFrame(ev, src, []*eval.Port{
		{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr},
	})
	fm.SetLocal(ns)
	err = fm.Eval(op)
	if err != nil {
		panic(err)
	}
}

// ReadCode reads input from the user.
func (ed *Editor) ReadCode() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}
