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

func startInsertRaw(ed *Editor) {
	ed.reader.SetRaw(true)
	ed.mode = rawInsert{}
}

func insertRaw(ed *Editor, r rune) {
	ed.insertAtDot(string(r))
	ed.reader.SetRaw(false)
	ed.mode = &ed.insert
}

func (rawInsert) Binding(map[string]eval.Variable, ui.Key) eval.CallableValue {
	// The raw insert mode does not handle keys.
	return nil
}

func (ri rawInsert) ModeLine() renderer {
	return modeLineRenderer{" RAW ", ""}
}
