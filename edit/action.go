package edit

import . "github.com/elves/elvish/edit/edtypes"

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
