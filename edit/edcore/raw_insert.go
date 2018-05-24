package edcore

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Raw insert mode is a special mode, in that it does not use the normal key
// binding. Rather, insertRaw is called directly from the main loop in
// Editor.ReadLine.

type rawInsert struct {
}

func (ed *editor) startInsertRaw() {
	ed.reader.SetRaw(true)
	ed.mode = rawInsert{}
}

func insertRaw(ed *editor, r rune) {
	ed.InsertAtDot(string(r))
	ed.reader.SetRaw(false)
	ed.SetModeInsert()
}

func (rawInsert) Teardown() {}

func (rawInsert) Binding(ui.Key) eval.Callable {
	// The raw insert mode does not handle keys.
	return nil
}

func (ri rawInsert) ModeLine() ui.Renderer {
	return ui.NewModeLineRenderer(" RAW ", "")
}
