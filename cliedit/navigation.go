package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/addons/navigation"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
)

//elvdoc:var selected-file
//
// Name of the currently selected file in navigation mode. $nil if not in
// navigation mode.

//elvdoc:var navigation:binding
//
// Keybinding for the navigation mode.

//elvdoc:fn navigation:start
//
// Start the navigation mode.

//elvdoc:fn navigation:insert-selected
//
// Inserts the selected filename.

func navInsertSelected(app cli.App) {
	insertAtDot(app, " "+parse.Quote(navigation.SelectedName(app)))
}

//elvdoc:fn navigation:insert-selected-and-quit
//
// Inserts the selected filename and closes the navigation addon.

func navInsertSelectedAndQuit(app cli.App) {
	navInsertSelected(app)
	closeListing(app)
}

//elvdoc:fn navigation:trigger-filter
//
// Toggles the filtering status of the navigation addon.

func navToggleFilter(app cli.App) {
	navigation.MutateFiltering(app, func(b bool) bool { return !b })
}

//elvdoc:fn navigation:trigger-shown-hidden
//
// Toggles whether the navigation addon should be showing hidden files.

func navToggleShowHidden(app cli.App) {
	navigation.MutateShowHidden(app, func(b bool) bool { return !b })
}

func initNavigation(app cli.App, ev *eval.Evaler, ns eval.Ns) {
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(app, ev, bindingVar)

	selectedFileVar := vars.FromGet(func() interface{} {
		name := navigation.SelectedName(app)
		if name == "" {
			return nil
		}
		return name
	})

	ns.Add("selected-file", selectedFileVar)
	ns.AddNs("navigation",
		eval.Ns{
			"binding": bindingVar,
		}.AddGoFns("<edit:navigation>", map[string]interface{}{
			"start": func() {
				navigation.Start(app, navigation.Config{Binding: binding})
			},
			"left":      func() { navigation.Ascend(app) },
			"right":     func() { navigation.Descend(app) },
			"up":        func() { navigation.Select(app, listbox.Prev) },
			"down":      func() { navigation.Select(app, listbox.Next) },
			"page-up":   func() { navigation.Select(app, listbox.PrevPage) },
			"page-down": func() { navigation.Select(app, listbox.NextPage) },

			"file-preview-up":   func() { navigation.ScrollPreview(app, -1) },
			"file-preview-down": func() { navigation.ScrollPreview(app, 1) },

			"insert-selected":          func() { navInsertSelected(app) },
			"insert-selected-and-quit": func() { navInsertSelectedAndQuit(app) },

			"trigger-filter":       func() { navToggleFilter(app) },
			"trigger-shown-hidden": func() { navToggleShowHidden(app) },
		}))
}
