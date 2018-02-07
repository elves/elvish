package edit

// Action is used in the SetAction method of the Editor to schedule a special
// Action after a keybinding has been executed.
type Action int

const (
	NoAction Action = iota
	// ReprocessKey makes the editor to reprocess the keybinding.
	ReprocessKey
	// CommitLine makes the editor return with the current line.
	CommitLine
	// CommitEOF makes the editor return with an EOF.
	CommitEOF
)

func (ed *Editor) SetAction(action Action) {
	if ed.nextAction == NoAction {
		ed.nextAction = action
	}
}

func (ed *Editor) popAction() Action {
	action := ed.nextAction
	ed.nextAction = NoAction
	return action
}
