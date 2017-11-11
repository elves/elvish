package edit

// action is used in the nextAction field of editorState to schedule a special
// action after a keybinding has been executed.
type action int

const (
	noAction action = iota
	// reprocessKey makes the editor to reprocess the keybinding.
	reprocessKey
	// commitLine makes the editor return with the current line.
	commitLine
	// commitEOF makes the editor return with an EOF.
	commitEOF
)

func (ed *Editor) setAction(action action) {
	if ed.nextAction == noAction {
		ed.nextAction = action
	}
}

func (ed *Editor) popAction() action {
	action := ed.nextAction
	ed.nextAction = noAction
	return action
}
