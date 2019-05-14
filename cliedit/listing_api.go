package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func initListing() (*bindingMap, eval.Ns) {
	binding := emptyBindingMap
	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
	}.AddGoFns("<edit:listing>:", map[string]interface{}{
		"up":               cli.ListingUp,
		"down":             cli.ListingDown,
		"up-cycle":         cli.ListingUpCycle,
		"down-cycle":       cli.ListingDownCycle,
		"toggle-filtering": cli.ListingToggleFiltering,
		"accept":           cli.ListingAccept,
		"accept-close":     cli.ListingAcceptClose,
		"default":          cli.ListingDefault,
	})
	return &binding, ns
}
