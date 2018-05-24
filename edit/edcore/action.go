package edcore

import "github.com/elves/elvish/edit/eddefs"

func (ed *editor) SetAction(action eddefs.Action) {
	if ed.nextAction == noAction {
		ed.nextAction = action
	}
}

func (ed *editor) popAction() eddefs.Action {
	action := ed.nextAction
	ed.nextAction = noAction
	return action
}
