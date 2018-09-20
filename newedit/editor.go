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
	ed.Config.Raw.Highlighter = highlight.Highlight

	ns := eval.NewNs().
		Add("max-height", vars.FromPtrWithMutex(
			&ed.Config.Raw.MaxHeight, &ed.Config.Mutex))

	ed.Config.Raw = core.RawConfig{
		Highlighter: highlight.Highlight,
		Prompt:      makePrompt(ed, ev, ns, defaultPrompt, "prompt"),
		RPrompt:     makePrompt(ed, ev, ns, defaultRPrompt, "rprompt"),
	}

	return &Editor{ed, ns}
}

func (ed *Editor) ReadLine() (string, error) {
	return ed.core.ReadCode()
}

func (ed *Editor) Ns() eval.Ns {
	return ed.ns
}

func (ed *Editor) Close() {}
