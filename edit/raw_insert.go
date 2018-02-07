package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

// Raw insert mode is a special mode, in that it does not use the normal key
// binding. Rather, insertRaw is called directly from the main loop in
// Editor.ReadLine.

type rawInsert struct {
}

func (ed *Editor) startInsertRaw() {
	ed.reader.SetRaw(true)
	ed.mode = rawInsert{}
}

func insertRaw(ed *Editor, r rune) {
	ed.insertAtDot(string(r))
	ed.reader.SetRaw(false)
	ed.SetModeInsert()
}

func (rawInsert) Binding(*Editor, ui.Key) eval.Callable {
	// The raw insert mode does not handle keys.
	return nil
}

func (ri rawInsert) ModeLine() ui.Renderer {
	return modeLineRenderer{" RAW ", ""}
}
