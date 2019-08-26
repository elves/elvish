package cliedit

import (
	"github.com/elves/elvish/cli/addons/histlist"
	"github.com/elves/elvish/cli/addons/lastcmd"
	"github.com/elves/elvish/cli/addons/location"
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func initListing() (*bindingMap, eval.Ns) {
	binding := emptyBindingMap
	ns := eval.Ns{
		"binding": vars.FromPtr(&binding),
	}.AddGoFns("<edit:listing>:", map[string]interface{}{
		/*
			"up":               cli.ListingUp,
			"down":             cli.ListingDown,
			"up-cycle":         cli.ListingUpCycle,
			"down-cycle":       cli.ListingDownCycle,
			"toggle-filtering": cli.ListingToggleFiltering,
			"accept":           cli.ListingAccept,
			"accept-close":     cli.ListingAcceptClose,
			"default":          cli.ListingDefault,
		*/
	})
	return &binding, ns
}

// Initializes states for the histlist mode and its API.
func initHistlist(app *cli.App, ev *eval.Evaler, lsBinding *bindingMap, store histutil.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:histlist>", "start", func() {
		histlist.Start(app, histlist.Config{binding, store})
	})
}

// Initializes states for the lastcmd mode and its API.
func initLastcmd(app *cli.App, ev *eval.Evaler, lsBinding *bindingMap, store histutil.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:lastcmd>", "start", func() {
		// TODO: Specify wordifier
		lastcmd.Start(app, lastcmd.Config{binding, store, nil})
	})
}

func initLocation(app *cli.App, ev *eval.Evaler, lsBinding *bindingMap, store location.Store) eval.Ns {
	m := emptyBindingMap
	binding := newMapBinding(app, ev, &m, lsBinding)
	return eval.Ns{
		"binding": vars.FromPtr(&m),
	}.AddGoFn("<edit:location>", "start", func() {
		location.Start(app, location.Config{binding, store})
	})
}
