package newedit

import (
	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/newedit/clitypes"
)

//elvdoc:fn binding-map
//
// Converts a normal map into a binding map.

var makeBindingMap = eddefs.MakeBindingMap

//elvdoc:fn reset-mode
//
// Resets the mode to the default mode.

func makeResetMode(st *clitypes.State) func() {
	return func() { st.SetMode(nil) }
}
