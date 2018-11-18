package newedit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/core"
	"github.com/elves/elvish/newedit/highlight"
	"github.com/elves/elvish/parse"
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
func NewEditor(in, out *os.File, ev *eval.Evaler) *Editor {
	ed := core.NewEditor(core.NewTTY(in, out), core.NewSignalSource())
	ed.Highlighter = highlight.Highlight

	ns := eval.NewNs().
		Add("max-height",
			vars.FromPtrWithMutex(&ed.Config.Raw.MaxHeight, &ed.Config.Mutex)).
		AddBuiltinFns("<edit>", map[string]interface{}{
			"binding-map":  makeBindingMap,
			"exit-binding": exitBinding,
			"commit-code":  commitCode,
			"commit-eof":   commitEOF,
		})

	// Hooks
	ns["before-readline"], ed.BeforeReadline = initBeforeReadline(ev)
	ns["after-readline"], ed.AfterReadline = initAfterReadline(ev)

	// Prompts
	ed.Prompt = makePrompt(ed, ev, ns, defaultPrompt, "prompt")
	ed.RPrompt = makePrompt(ed, ev, ns, defaultRPrompt, "rprompt")

	// Insert mode
	insertMode, insertNs := initInsert(ed, ev)
	ed.InitMode = insertMode
	ns.AddNs("insert", insertNs)

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
