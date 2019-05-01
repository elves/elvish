package newedit

import (
	"github.com/elves/elvish/cli/listing"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func initListing(a app) (*listing.Mode, *bindingMap, eval.Ns) {
	mode := &listing.Mode{}
	binding := emptyBindingMap
	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
	}.AddGoFns("<edit:listing>:", map[string]interface{}{
		"up":         func() { mode.MutateStates((*listing.State).Up) },
		"down":       func() { mode.MutateStates((*listing.State).Down) },
		"up-cycle":   func() { mode.MutateStates((*listing.State).UpCycle) },
		"down-cycle": func() { mode.MutateStates((*listing.State).DownCycle) },

		"toggle-filtering": func() { mode.MutateStates((*listing.State).ToggleFiltering) },

		"accept":       func() { mode.AcceptItem(a.State()) },
		"accept-close": func() { mode.AcceptItemAndClose(a.State()) },

		"default": func() { mode.DefaultHandler(a.State()) },
	})
	return mode, &binding, ns
}
