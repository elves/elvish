package edit

func insertRaw(ed *Editor) {
	ed.reader.SetRaw(true)
	ed.mode = rawInsert{}
}

type rawInsert struct {
}

func (ri rawInsert) Mode() ModeType {
	return modeRawInsert
}

func (ri rawInsert) ModeLine() renderer {
	return modeLineRenderer{" RAW ", ""}
}
