package cliedit

import (
	"os"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/histlist"
	"github.com/elves/elvish/cli/addons/lastcmd"
	"github.com/elves/elvish/cli/addons/location"
	"github.com/elves/elvish/cli/el/combobox"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/histutil"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/store/storedefs"
	"github.com/xiaq/persistent/hashmap"
)

func initListings(app cli.App, ev *eval.Evaler, ns eval.Ns, st storedefs.Store, fuser *histutil.Fuser) {
	bindingVar := newBindingVar(emptyBindingMap)
	ns.AddNs("listing",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:listing>:", map[string]interface{}{
			"accept":       func() { listingAccept(app) },
			"accept-close": func() { listingAcceptClose(app) },
			"close":        func() { closeListing(app) },
			"up":           func() { listingUp(app) },
			"down":         func() { listingDown(app) },
			"up-cycle":     func() { listingUpCycle(app) },
			"down-cycle":   func() { listingDownCycle(app) },
			"page-up":      func() { listingPageUp(app) },
			"page-down":    func() { listingPageDown(app) },
			/*
				"toggle-filtering": cli.ListingToggleFiltering,
			*/
		}))

	var histStore histutil.Store
	if fuser != nil {
		histStore = fuserWrapper{fuser}
	}

	initHistlist(app, ev, ns, histStore, bindingVar)
	initLastcmd(app, ev, ns, histStore, bindingVar)
	initLocation(app, ev, ns, st, bindingVar)
}

func initHistlist(app cli.App, ev *eval.Evaler, ns eval.Ns, histStore histutil.Store, commonBindingVar vars.PtrVar) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar, commonBindingVar)
	dedup := newBoolVar(true)
	caseSensitive := newBoolVar(true)
	ns.AddNs("histlist",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:histlist>", map[string]interface{}{
			"start": func() {
				histlist.Start(app, histlist.Config{
					Binding: binding, Store: histStore,
					CaseSensitive: func() bool {
						return caseSensitive.Get().(bool)
					},
					Dedup: func() bool {
						return dedup.Get().(bool)
					},
				})
			},
			"toggle-case-sensitivity": func() {
				caseSensitive.Set(!caseSensitive.Get().(bool))
				listingRefilter(app)
				app.Redraw()
			},
			"toggle-dedup": func() {
				dedup.Set(!dedup.Get().(bool))
				listingRefilter(app)
				app.Redraw()
			},
		}))
}

func initLastcmd(app cli.App, ev *eval.Evaler, ns eval.Ns, histStore histutil.Store, commonBindingVar vars.PtrVar) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar, commonBindingVar)
	ns.AddNs("lastcmd",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFn("<edit:lastcmd>", "start", func() {
			// TODO: Specify wordifier
			lastcmd.Start(app, lastcmd.Config{
				Binding: binding, Store: histStore})
		}))
}

func initLocation(app cli.App, ev *eval.Evaler, ns eval.Ns, st storedefs.Store, commonBindingVar vars.PtrVar) {
	bindingVar := newBindingVar(emptyBindingMap)
	pinnedVar := newListVar(vals.EmptyList)
	hiddenVar := newListVar(vals.EmptyList)
	workspacesVar := newMapVar(vals.EmptyMap)

	binding := newMapBinding(app, ev, bindingVar, commonBindingVar)
	workspaceIterator := location.WorkspaceIterator(
		adaptToIterateStringPair(workspacesVar))

	ns.AddNs("location",
		eval.Ns{
			"binding":    bindingVar,
			"hidden":     hiddenVar,
			"pinned":     pinnedVar,
			"workspaces": workspacesVar,
		}.AddGoFn("<edit:location>", "start", func() {
			location.Start(app, location.Config{
				Binding: binding, Store: dirStore{ev, st},
				IteratePinned:     adaptToIterateString(pinnedVar),
				IterateHidden:     adaptToIterateString(hiddenVar),
				IterateWorkspaces: workspaceIterator,
			})
		}))
	ev.AddAfterChdir(func(string) {
		wd, err := os.Getwd()
		if err != nil {
			// TODO(xiaq): Surface the error.
			return
		}
		st.AddDir(wd, 1)
		kind, root := workspaceIterator.Parse(wd)
		if kind != "" {
			st.AddDir(kind+wd[len(root):], 1)
		}
	})
}

//elvdoc:fn listing:accept
//
// Accepts the current selected listing item.

func listingAccept(app cli.App) {
	w, ok := app.CopyState().Addon.(combobox.Widget)
	if !ok {
		return
	}
	w.CodeArea().Submit()
}

//elvdoc:fn listing:accept-close
//
// Accepts the current selected listing item and closes the listing.

func listingAcceptClose(app cli.App) {
	listingAccept(app)
	closeListing(app)
}

//elvdoc:fn listing:up
//
// Moves the cursor up in listing mode.

func listingUp(app cli.App) { listingSelect(app, listbox.Prev) }

//elvdoc:fn listing:down
//
// Moves the cursor down in listing mode.

func listingDown(app cli.App) { listingSelect(app, listbox.Next) }

//elvdoc:fn listing:up-cycle
//
// Moves the cursor up in listing mode, or to the last item if the first item is
// currently selected.

func listingUpCycle(app cli.App) { listingSelect(app, listbox.PrevWrap) }

//elvdoc:fn listing:down-cycle
//
// Moves the cursor down in listing mode, or to the first item if the last item is
// currently selected.

func listingDownCycle(app cli.App) { listingSelect(app, listbox.NextWrap) }

//elvdoc:fn listing:page-up
// Moves the cursor up one page.

func listingPageUp(app cli.App) { listingSelect(app, listbox.PrevPage) }

//elvdoc:fn listing:page-down
// Moves the cursor down one page.

func listingPageDown(app cli.App) { listingSelect(app, listbox.NextPage) }

//elvdoc:fn listing:left
// Moves the cursor left in listing mode.

func listingLeft(app cli.App) { listingSelect(app, listbox.Left) }

//elvdoc:fn listing:right
//
// Moves the cursor right in listing mode.

func listingRight(app cli.App) { listingSelect(app, listbox.Right) }

func listingSelect(app cli.App, f func(listbox.State) int) {
	w, ok := app.CopyState().Addon.(combobox.Widget)
	if !ok {
		return
	}
	w.ListBox().Select(f)
}

func listingRefilter(app cli.App) {
	w, ok := app.CopyState().Addon.(combobox.Widget)
	if !ok {
		return
	}
	w.Refilter()
}

//elvdoc:var location:hidden
//
// A list of directories to hide in the location addon.

//elvdoc:var location:pinned
//
// A list of directories to always show at the top of the list of the location
// addon.

//elvdoc:var location:workspaces
//
// A map mapping types of workspaces to their patterns.

func adaptToIterateString(variable vars.Var) func(func(string)) {
	return func(f func(s string)) {
		vals.Iterate(variable.Get(), func(v interface{}) bool {
			f(vals.ToString(v))
			return true
		})
	}
}

func adaptToIterateStringPair(variable vars.Var) func(func(string, string) bool) {
	return func(f func(a, b string) bool) {
		m := variable.Get().(hashmap.Map)
		for it := m.Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			ks, kok := k.(string)
			vs, vok := v.(string)
			if kok && vok {
				next := f(ks, vs)
				if !next {
					break
				}
			}
		}
	}
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
	st storedefs.Store
}

func (d dirStore) Chdir(path string) error {
	return d.ev.Chdir(path)
}

func (d dirStore) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	return d.st.Dirs(blacklist)
}

func (d dirStore) Getwd() (string, error) {
	return os.Getwd()
}
