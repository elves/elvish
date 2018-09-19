package newedit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/core"
	"github.com/elves/elvish/newedit/highlight"
)

// Editor is the line editor for Elvish.
//
// This currently implements the same interface as *Editor in the old edit
// package to ease transition. TODO: Rename ReadLine to ReadCode and remove
// Close.
type Editor interface {
	ReadLine() (string, error)
	Ns() eval.Ns
	Close()
}

type editor struct {
	core *core.Editor
	ns   eval.Ns
}

// NewEditor creates a new editor from input and output terminal files.
func NewEditor(in, out *os.File) Editor {
	ed := core.NewEditor(core.NewTTY(in, out), core.NewSignalSource())
	ed.Config.Raw.Highlighter = highlight.Highlight

	ns := eval.NewNs().
		Add("max-height", vars.FromPtrWithMutex(
			&ed.Config.Raw.MaxHeight, &ed.Config.Mutex))

	return &editor{ed, ns}
}

func (ed *editor) ReadLine() (string, error) {
	return ed.core.ReadCode()
}

func (ed *editor) Ns() eval.Ns {
	return ed.ns
}

func (ed *editor) Close() {}
