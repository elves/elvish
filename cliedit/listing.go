package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/histlist"
	"github.com/elves/elvish/cli/addons/lastcmd"
	"github.com/elves/elvish/cli/addons/location"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/store/storedefs"
)

func initListings(app cli.App, ev *eval.Evaler, ns eval.Ns, st storedefs.Store, fuser *histutil.Fuser) {
	var histStore histutil.Store
	if fuser != nil {
		histStore = fuserWrapper{fuser}
	}
	dirStore := dirStore{ev}

	// Common binding and the listing: module.
	lsMap := newBindingVar(emptyBindingMap)
	ns.AddNs("listing",
		eval.Ns{
			"binding": lsMap,
		}.AddGoFns("<edit:listing>:", map[string]interface{}{
			"close":      func() { closeListing(app) },
			"up":         func() { listingUp(app) },
			"down":       func() { listingDown(app) },
			"up-cycle":   func() { listingUpCycle(app) },
			"down-cycle": func() { listingDownCycle(app) },
			/*
				"toggle-filtering": cli.ListingToggleFiltering,
				"accept":           cli.ListingAccept,
				"accept-close":     cli.ListingAcceptClose,
			*/
		}))

	histlistMap := newBindingVar(emptyBindingMap)
	histlistBinding := newMapBinding(app, ev, histlistMap, lsMap)
	ns.AddNs("histlist",
		eval.Ns{
			"binding": histlistMap,
		}.AddGoFn("<edit:histlist>", "start", func() {
			histlist.Start(app, histlist.Config{histlistBinding, histStore})
		}))

	lastcmdMap := newBindingVar(emptyBindingMap)
	lastcmdBinding := newMapBinding(app, ev, lastcmdMap, lsMap)
	ns.AddNs("lastcmd",
		eval.Ns{
			"binding": lastcmdMap,
		}.AddGoFn("<edit:lastcmd>", "start", func() {
			// TODO: Specify wordifier
			lastcmd.Start(app, lastcmd.Config{lastcmdBinding, histStore, nil})
		}))

	locationMap := newBindingVar(emptyBindingMap)
	locationBinding := newMapBinding(app, ev, locationMap, lsMap)
	ns.AddNs("location",
		eval.Ns{
			"binding": locationMap,
		}.AddGoFn("<edit:location>", "start", func() {
			location.Start(app, location.Config{locationBinding, dirStore})
		}))
}

//elvdoc:fn listing:up
//
// Moves the cursor up in listing mode.

//elvdoc:fn listing:down
//
// Moves the cursor down in listing mode.

//elvdoc:fn listing:up-cycle
//
// Moves the cursor up in listing mode, or to the last item if the first item is
// currently selected.

//elvdoc:fn listing:down-cycle
//
// Moves the cursor down in listing mode, or to the first item if the last item is
// currently selected.

func listingUp(app cli.App)        { listingSelect(app, listbox.Prev) }
func listingDown(app cli.App)      { listingSelect(app, listbox.Next) }
func listingUpCycle(app cli.App)   { listingSelect(app, listbox.PrevWrap) }
func listingDownCycle(app cli.App) { listingSelect(app, listbox.NextWrap) }

func listingSelect(app cli.App, f func(listbox.State) int) {
	w, ok := app.CopyAppState().Listing.(combobox.Widget)
	if !ok {
		return
	}
	w.ListBox().Select(f)
}

// Wraps the histutil.Fuser interface to implement histutil.Store. This is a
// bandaid as we cannot change the implementation of Fuser without breaking its
// other users. Eventually Fuser should implement Store directly.
type fuserWrapper struct {
	*histutil.Fuser
}

func (f fuserWrapper) AddCmd(cmd histutil.Entry) (int, error) {
	return f.Fuser.AddCmd(cmd.Text)
}

// Wraps an Evaler to implement the cli.DirStore interface.
type dirStore struct {
	ev *eval.Evaler
}

func (d dirStore) Chdir(path string) error {
	return d.ev.Chdir(path)
}

func (d dirStore) Dirs() ([]storedefs.Dir, error) {
	return d.ev.DaemonClient.Dirs(map[string]struct{}{})
}
