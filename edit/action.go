package edit

import "github.com/elves/elvish/edit/edtypes"

func (ed *editor) SetAction(action edtypes.Action) {
	if ed.nextAction == noAction {
		ed.nextAction = action
	}
}

func (ed *editor) popAction() edtypes.Action {
	action := ed.nextAction
	ed.nextAction = noAction
	return action
}
