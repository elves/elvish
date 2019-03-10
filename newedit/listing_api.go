package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/listing"
)

func initListing() (*listing.Mode, *BindingMap, eval.Ns) {
	mode := &listing.Mode{}
	binding := EmptyBindingMap
	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
	}.AddBuiltinFns("<edit:listing>:", map[string]interface{}{
		"up":         func() { mode.MutateStates((*listing.State).Up) },
		"down":       func() { mode.MutateStates((*listing.State).Down) },
		"up-cycle":   func() { mode.MutateStates((*listing.State).UpCycle) },
		"down-cycle": func() { mode.MutateStates((*listing.State).DownCycle) },
	})
	return mode, &binding, ns
}
