package edit

import (
	"os"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/histutil"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/edit/filter"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/store/storedefs"
)

func initListings(ed *Editor, ev *eval.Evaler, st storedefs.Store, histStore histutil.Store, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	app := ed.app
	nb.AddNs("listing",
		eval.BuildNsNamed("edit:listing").
			AddVar("binding", bindingVar).
			AddGoFns(map[string]interface{}{
				"accept":     func() { listingAccept(app) },
				"up":         func() { listingUp(app) },
				"down":       func() { listingDown(app) },
				"up-cycle":   func() { listingUpCycle(app) },
				"down-cycle": func() { listingDownCycle(app) },
				"page-up":    func() { listingPageUp(app) },
				"page-down":  func() { listingPageDown(app) },
				"start-custom": func(fm *eval.Frame, opts customListingOpts, items interface{}) {
					listingStartCustom(ed, fm, opts, items)
				},
			}))

	initHistlist(ed, ev, histStore, bindingVar, nb)
	initLastcmd(ed, ev, histStore, bindingVar, nb)
	initLocation(ed, ev, st, bindingVar, nb)
}

var filterSpec = modes.FilterSpec{
	Maker: func(f string) func(string) bool {
		q, _ := filter.Compile(f)
		if q == nil {
			return func(string) bool { return true }
		}
		return q.Match
	},
	Highlighter: filter.Highlight,
}

func initHistlist(ed *Editor, ev *eval.Evaler, histStore histutil.Store, commonBindingVar vars.PtrVar, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar, commonBindingVar)
	dedup := newBoolVar(true)
	nb.AddNs("histlist",
		eval.BuildNsNamed("edit:histlist").
			AddVar("binding", bindingVar).
			AddGoFns(map[string]interface{}{
				"start": func() {
					w, err := modes.NewHistlist(ed.app, modes.HistlistSpec{
						Bindings: bindings,
						AllCmds:  histStore.AllCmds,
						Dedup: func() bool {
							return dedup.Get().(bool)
						},
						Filter: filterSpec,
					})
					startMode(ed.app, w, err)
				},
				"toggle-dedup": func() {
					dedup.Set(!dedup.Get().(bool))
					listingRefilter(ed.app)
					ed.app.Redraw()
				},
			}))
}

func initLastcmd(ed *Editor, ev *eval.Evaler, histStore histutil.Store, commonBindingVar vars.PtrVar, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar, commonBindingVar)
	nb.AddNs("lastcmd",
		eval.BuildNsNamed("edit:lastcmd").
			AddVar("binding", bindingVar).
			AddGoFn("start", func() {
				// TODO: Specify wordifier
				w, err := modes.NewLastcmd(ed.app, modes.LastcmdSpec{
					Bindings: bindings, Store: histStore})
				startMode(ed.app, w, err)
			}))
}

func initLocation(ed *Editor, ev *eval.Evaler, st storedefs.Store, commonBindingVar vars.PtrVar, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	pinnedVar := newListVar(vals.EmptyList)
	hiddenVar := newListVar(vals.EmptyList)
	workspacesVar := newMapVar(vals.EmptyMap)

	bindings := newMapBindings(ed, ev, bindingVar, commonBindingVar)
	workspaceIterator := modes.LocationWSIterator(
		adaptToIterateStringPair(workspacesVar))

	nb.AddNs("location",
		eval.BuildNsNamed("edit:location").
			AddVars(map[string]vars.Var{
				"binding":    bindingVar,
				"hidden":     hiddenVar,
				"pinned":     pinnedVar,
				"workspaces": workspacesVar,
			}).
			AddGoFn("start", func() {
				w, err := modes.NewLocation(ed.app, modes.LocationSpec{
					Bindings: bindings, Store: dirStore{ev, st},
					IteratePinned:     adaptToIterateString(pinnedVar),
					IterateHidden:     adaptToIterateString(hiddenVar),
					IterateWorkspaces: workspaceIterator,
					Filter:            filterSpec,
				})
				startMode(ed.app, w, err)
			}))
	ev.AddAfterChdir(func(string) {
		wd, err := os.Getwd()
		if err != nil {
			// TODO(xiaq): Surface the error.
			return
		}
		if st != nil {
			st.AddDir(wd, 1)
			kind, root := workspaceIterator.Parse(wd)
			if kind != "" {
				st.AddDir(kind+wd[len(root):], 1)
			}
		}
	})
}

//elvdoc:fn listing:accept
//
// Accepts the current selected listing item.

func listingAccept(app cli.App) {
	if w, ok := activeComboBox(app); ok {
		w.ListBox().Accept()
	}
}

//elvdoc:fn listing:up
//
// Moves the cursor up in listing mode.

func listingUp(app cli.App) { listingSelect(app, tk.Prev) }

//elvdoc:fn listing:down
//
// Moves the cursor down in listing mode.

func listingDown(app cli.App) { listingSelect(app, tk.Next) }

//elvdoc:fn listing:up-cycle
//
// Moves the cursor up in listing mode, or to the last item if the first item is
// currently selected.

func listingUpCycle(app cli.App) { listingSelect(app, tk.PrevWrap) }

//elvdoc:fn listing:down-cycle
//
// Moves the cursor down in listing mode, or to the first item if the last item is
// currently selected.

func listingDownCycle(app cli.App) { listingSelect(app, tk.NextWrap) }

//elvdoc:fn listing:page-up
//
// Moves the cursor up one page.

func listingPageUp(app cli.App) { listingSelect(app, tk.PrevPage) }

//elvdoc:fn listing:page-down
//
// Moves the cursor down one page.

func listingPageDown(app cli.App) { listingSelect(app, tk.NextPage) }

//elvdoc:fn listing:left
//
// Moves the cursor left in listing mode.

func listingLeft(app cli.App) { listingSelect(app, tk.Left) }

//elvdoc:fn listing:right
//
// Moves the cursor right in listing mode.

func listingRight(app cli.App) { listingSelect(app, tk.Right) }

func listingSelect(app cli.App, f func(tk.ListBoxState) int) {
	if w, ok := activeComboBox(app); ok {
		w.ListBox().Select(f)
	}
}

func listingRefilter(app cli.App) {
	if w, ok := activeComboBox(app); ok {
		w.Refilter()
	}
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
		m := variable.Get().(vals.Map)
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

// Wraps an Evaler to implement the cli.DirStore interface.
type dirStore struct {
	ev *eval.Evaler
	st storedefs.Store
}

func (d dirStore) Chdir(path string) error {
	return d.ev.Chdir(path)
}

func (d dirStore) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	if d.st == nil {
		// A "no daemon" build won't have have a storedefs.Store object.
		// Fail gracefully rather than panic.
		return []storedefs.Dir{}, nil
	}
	return d.st.Dirs(blacklist)
}

func (d dirStore) Getwd() (string, error) {
	return os.Getwd()
}

func startMode(app cli.App, w tk.Widget, err error) {
	if w != nil {
		app.PushAddon(w)
		app.Redraw()
	}
	if err != nil {
		app.Notify(modes.ErrorText(err))
	}
}

func activeComboBox(app cli.App) (tk.ComboBox, bool) {
	w, ok := app.ActiveWidget().(tk.ComboBox)
	return w, ok
}
