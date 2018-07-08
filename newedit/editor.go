package newedit

import (
	"os"

	"github.com/elves/elvish/newedit/core"
)

// Editor is the line editor for Elvish.
//
// This currently implements the same interface as *Editor in the old edit
// package to ease transition. TODO: Rename ReadLine to ReadCode and remove
// Close.
type Editor interface {
	ReadLine() (string, error)
}

type editor struct {
	core *core.Editor
}

func NewEditor(in, out *os.File) Editor {
	ed := core.NewEditor(core.NewTTY(in, out))
	return &editor{ed}
}

func (ed *editor) ReadLine() (string, error) {
	return ed.core.ReadCode()
}

func (ed *editor) Close() {}
