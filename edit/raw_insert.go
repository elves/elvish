package edit

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

func (ri rawInsert) Mode() ModeType {
	return modeRawInsert
}

func (ri rawInsert) ModeLine() renderer {
	return modeLineRenderer{" RAW ", ""}
}
