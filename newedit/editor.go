package newedit

import (
	"fmt"
	"os"

	"github.com/elves/elvish/cli/clicore"
	"github.com/elves/elvish/edit/history/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/highlight"
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
	app *clicore.App
	ns  eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File, ev *eval.Evaler, st storedefs.Store) *Editor {
	app := clicore.NewAppFromFiles(in, out)

	app.Highlighter = highlight.NewHighlighter(
		highlight.Dep{Check: makeCheck(ev), HasCommand: makeHasCommand(ev)})

	ns := eval.NewNs().
		Add("max-height",
			vars.FromPtrWithMutex(&app.Config.Raw.MaxHeight, &app.Config.Mutex)).
		AddGoFns("<edit>", map[string]interface{}{
			"binding-map":  makeBindingMap,
			"exit-binding": exitBinding,
			"commit-code":  commitCode,
			"commit-eof":   commitEOF,
			"reset-mode":   makeResetMode(app.State()),
		}).
		AddGoFns("<edit>", bufferBuiltins(app.State()))

	histFuser, err := histutil.NewFuser(st)
	if err == nil {
		// Add the builtin hook of appending history in after-readline.
		app.AddAfterReadline(func(code string) {
			err := histFuser.AddCmd(code)
			if err != nil {
				fmt.Fprintln(out, "failed to add command to history")
			}
		})
	} else {
		fmt.Fprintln(out, "failed to initialize history facilities")
	}

	// Elvish hook APIs
	var beforeReadline func()
	ns["before-readline"], beforeReadline = initBeforeReadline(ev)
	app.AddBeforeReadline(beforeReadline)
	var afterReadline func(string)
	ns["after-readline"], afterReadline = initAfterReadline(ev)
	app.AddAfterReadline(afterReadline)

	// Prompts
	app.Prompt = makePrompt(app, ev, ns, defaultPrompt, "prompt")
	app.RPrompt = makePrompt(app, ev, ns, defaultRPrompt, "rprompt")

	// Insert mode
	insertMode, insertNs := initInsert(app, ev)
	app.InitMode = insertMode
	ns.AddNs("insert", insertNs)

	// Listing modes.
	lsMode, lsBinding, lsNs := initListing(app)
	ns.AddNs("listing", lsNs)

	lastcmdNs := initLastcmd(app, ev, st, lsMode, lsBinding)
	ns.AddNs("lastcmd", lastcmdNs)

	histlistNs := initHistlist(app, ev, histFuser.AllCmds, lsMode, lsBinding)
	ns.AddNs("histlist", histlistNs)

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
