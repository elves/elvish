package newedit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/core"
	"github.com/elves/elvish/newedit/highlight"
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
		AddFn("binding-map",
			eval.NewBuiltinFn("<edit>:binding-map", makeBindingMap)).
		AddFn("exit-binding",
			eval.NewBuiltinFn("<edit>:exit-binding", exitBinding)).
		AddFn("commit-code",
			eval.NewBuiltinFn("<edit>:commit-code", commitCode))

	ns["before-readline"], ed.BeforeReadline = initBeforeReadline(ev)
	ns["after-readline"], ed.AfterReadline = initAfterReadline(ev)

	ed.Prompt = makePrompt(ed, ev, ns, defaultPrompt, "prompt")
	ed.RPrompt = makePrompt(ed, ev, ns, defaultRPrompt, "rprompt")

	// TODO: Initialize insert mode

	return &Editor{ed, ns}
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
