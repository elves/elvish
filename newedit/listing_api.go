package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/listing"
)

func initListing(ed editor) (*listing.Mode, *bindingMap, eval.Ns) {
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

		"accept":       func() { mode.AcceptItem(ed.State()) },
		"accept-close": func() { mode.AcceptItemAndClose(ed.State()) },

		"default": func() { mode.DefaultHandler(ed.State()) },
	})
	return mode, &binding, ns
}
