package newedit

import (
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/newedit/types"
	"github.com/elves/elvish/newedit/utils"
)

//elvish:doc-fn binding-map
//
// Converts a normal map into a binding map.

var makeBindingMap = eddefs.MakeBindingMap

//elvish:doc-fn exit-binding
//
// Exits the current binding handler. Internally, this works by raising a
// special exception.

func exitBinding() error {
	return utils.ActionError(types.NoAction)
}

//elvish:doc-fn commit-code
//
// Causes the current editor session to terminate, committing the code that is
// being edited. Internally, this works by raising a special exception.

func commitCode() error {
	return utils.ActionError(types.CommitCode)
}
