// Package edit implements the line editor for Elvish.
//
// The line editor is based on the cli package, which implements a general,
// Elvish-agnostic line editor, and multiple "addon" packages. This package
// glues them together and provides Elvish bindings for them.
package cliedit

import (
	"os"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store/storedefs"
)

// Editor is the interface line editor for Elvish.
//
// This currently implements the same interface as *Editor in the old edit
// package to ease transition.
//
// TODO: Rename ReadLine to ReadCode and remove Close.
type Editor struct {
	app cli.App
	ns  eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(tty cli.TTY, ev *eval.Evaler, st storedefs.Store) *Editor {
	ns := eval.NewNs()
	appSpec := cli.AppSpec{TTY: tty}

	fuser, err := histutil.NewFuser(st)
	if err != nil {
		// TODO(xiaq): Report the error.
	}

	if fuser != nil {
		appSpec.AfterReadline = []func(string){func(code string) {
			fuser.AddCmd(code)
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

	initListings(app, ev, ns, st, fuser)
	initNavigation(app, ev, ns)
	initCompletion(app, ev, ns)
	initHistWalk(app, ev, ns, fuser)

	initBufferBuiltins(app, ns)
	initDumpBuf(tty, ns)
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

// ReadLine reads input from the user.
func (ed *Editor) ReadLine() (string, error) {
	return ed.app.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}

// Close is a no-op.
func (ed *Editor) Close() {}
