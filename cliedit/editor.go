package cliedit

import (
	"fmt"
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

// Wraps the histutil.Fuser interface to implement histutil.Store. This is a
// bandaid as we cannot change the implementation of Fuser without breaking its
// other users. Eventually Fuser should implement Store directly.
type fuserWrapper struct {
	*histutil.Fuser
}

func (f fuserWrapper) AddCmd(cmd histutil.Entry) (int, error) {
	return f.Fuser.AddCmd(cmd.Text)
}

// Wraps an Evaler to implement the cli.DirStore interface.
type dirStore struct {
	ev *eval.Evaler
}

func (d dirStore) Chdir(path string) error {
	return d.ev.Chdir(path)
}

func (d dirStore) Dirs() ([]storedefs.Dir, error) {
	return d.ev.DaemonClient.Dirs(map[string]struct{}{})
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File, ev *eval.Evaler, st storedefs.Store) *Editor {
	ns := eval.NewNs()
	app := cli.NewApp(cli.NewTTY(in, out))

	initAPI(app, ev, ns)
	initPrompts(app, ev, ns)
	app.Config.Highlighter = makeHighlighter(ev)

	// Listing modes.
	lsBinding, lsNs := initListing()
	ns.AddNs("listing", lsNs)

	var histStore histutil.Store
	histFuser, err := histutil.NewFuser(st)
	if err == nil {
		histStore = fuserWrapper{histFuser}
	} else {
		fmt.Fprintln(out, "failed to initialize history facilities")
	}

	histlistNs := initHistlist(app, ev, lsBinding, histStore)
	ns.AddNs("histlist", histlistNs)

	lastcmdNs := initLastcmd(app, ev, lsBinding, histStore)
	ns.AddNs("lastcmd", lastcmdNs)

	dirStore := dirStore{ev}

	locationNs := initLocation(app, ev, lsBinding, dirStore)
	ns.AddNs("location", locationNs)

	// Evaluate default bindings.
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
