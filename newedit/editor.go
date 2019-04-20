package newedit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/core"
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
	core *core.Editor
	ns   eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File, ev *eval.Evaler, st storedefs.Store) *Editor {
	ed := core.NewEditor(core.NewTTY(in, out), core.NewSignalSource())

	ed.Highlighter = highlight.NewHighlighter(
		highlight.Dep{Check: makeCheck(ev), HasCommand: makeHasCommand(ev)})

	ns := eval.NewNs().
		Add("max-height",
			vars.FromPtrWithMutex(&ed.Config.Raw.MaxHeight, &ed.Config.Mutex)).
		AddGoFns("<edit>", map[string]interface{}{
			"binding-map":  makeBindingMap,
			"exit-binding": exitBinding,
			"commit-code":  commitCode,
			"commit-eof":   commitEOF,
			"reset-mode":   makeResetMode(ed.State()),
		}).
		AddGoFns("<edit>", bufferBuiltins(ed.State()))

	// Add the builtin hook of appending history in after-readline.
	ed.AddAfterReadline(func(code string) {
		st.AddCmd(code)
		// TODO: Log errors
	})

	// Elvish hook APIs
	var beforeReadline func()
	ns["before-readline"], beforeReadline = initBeforeReadline(ev)
	ed.AddBeforeReadline(beforeReadline)
	var afterReadline func(string)
	ns["after-readline"], afterReadline = initAfterReadline(ev)
	ed.AddAfterReadline(afterReadline)

	// Prompts
	ed.Prompt = makePrompt(ed, ev, ns, defaultPrompt, "prompt")
	ed.RPrompt = makePrompt(ed, ev, ns, defaultRPrompt, "rprompt")

	// Insert mode
	insertMode, insertNs := initInsert(ed, ev)
	ed.InitMode = insertMode
	ns.AddNs("insert", insertNs)

	// Listing modes.
	lsMode, lsBinding, lsNs := initListing(ed)
	ns.AddNs("listing", lsNs)
	lastcmdNs := initLastcmd(ed, ev, st, lsMode, lsBinding)
	ns.AddNs("lastcmd", lastcmdNs)

	// Evaluate default bindings.
	evalDefaultBinding(ev, ns)

	return &Editor{ed, ns}
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
	return ed.core.ReadCode()
}

// Ns returns a namespace for manipulating the editor from Elvish code.
func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}

// Close is a no-op.
func (ed *Editor) Close() {}
