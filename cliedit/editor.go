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
	app *cli.App
	ns  eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File, ev *eval.Evaler, st storedefs.Store) *Editor {
	ns := eval.NewNs()
	app := cli.NewApp(cli.NewTTY(in, out))

	fuser, err := histutil.NewFuser(st)
	if err != nil {
		// TODO(xiaq): Report the error.
	}

	initHighlighter(app, ev)
	initAPI(app, ev, ns)
	initPrompts(app, ev, ns)
	initListings(app, ev, ns, st, fuser)
	initNavigation(app, ev, ns)
	initCompletion(app, ev, ns)
	initHistWalk(app, ev, ns, fuser)
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
