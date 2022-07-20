package edit

import (
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

//elvdoc:var navigation:selected-file
//
// Name of the currently selected file in navigation mode. $nil if not in
// navigation mode.

//elvdoc:var navigation:binding
//
// ```elvish
// edit:navigation:binding
// ```
//
// Keybinding for the navigation mode.

//elvdoc:fn navigation:start
//
// ```elvish
// edit:navigation:start
// ```
//
// Start the navigation mode.

//elvdoc:fn navigation:insert-selected
//
// ```elvish
// edit:navigation:insert-selected
// ```
//
// Inserts the selected filename.

func navInsertSelected(app cli.App) {
	w, ok := activeNavigation(app)
	if !ok {
		return
	}
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	fname := w.SelectedName()
	if fname == "" {
		// User pressed Alt-Enter or Enter in an empty directory with nothing
		// selected; don't do anything.
		return
	}

	codeArea.MutateState(func(s *tk.CodeAreaState) {
		dot := s.Buffer.Dot
		if dot != 0 && !strings.ContainsRune(" \n", rune(s.Buffer.Content[dot-1])) {
			// The dot is not at the beginning of a buffer, and the previous
			// character is not a space or newline. Insert a space.
			s.Buffer.InsertAtDot(" ")
		}
		// Insert the selected filename.
		s.Buffer.InsertAtDot(parse.Quote(fname))
	})
}

//elvdoc:fn navigation:insert-selected-and-quit
//
// ```elvish
// edit:navigation:insert-selected-and-quit
// ```
//
// Inserts the selected filename and closes the navigation addon.

func navInsertSelectedAndQuit(app cli.App) {
	navInsertSelected(app)
	closeMode(app)
}

//elvdoc:fn navigation:trigger-filter
//
// ```elvish
// edit:navigation:trigger-filter
// ```
//
// Toggles the filtering status of the navigation addon.

//elvdoc:fn navigation:trigger-shown-hidden
//
// ```elvish
// edit:navigation:trigger-shown-hidden
// ```
//
// Toggles whether the navigation addon should be showing hidden files.

//elvdoc:var navigation:width-ratio
//
// A list of 3 integers, used for specifying the width ratio of the 3 columns in
// navigation mode.

func convertNavWidthRatio(v any) [3]int {
	var (
		numbers []int
		hasErr  bool
	)
	vals.Iterate(v, func(elem any) bool {
		var i int
		err := vals.ScanToGo(elem, &i)
		if err != nil {
			hasErr = true
			return false
		}
		numbers = append(numbers, i)
		return true
	})
	if hasErr || len(numbers) != 3 {
		// TODO: Handle the error.
		return [3]int{1, 3, 4}
	}
	var ret [3]int
	copy(ret[:], numbers)
	return ret
}

func initNavigation(ed *Editor, ev *eval.Evaler, nb eval.NsBuilder) {
	bindingVar := newBindingVar(emptyBindingsMap)
	bindings := newMapBindings(ed, ev, bindingVar)
	widthRatioVar := newListVar(vals.MakeList(1.0, 3.0, 4.0))

	selectedFileVar := vars.FromGet(func() any {
		if w, ok := activeNavigation(ed.app); ok {
			return w.SelectedName()
		}
		return nil
	})

	app := ed.app
	nb.AddVar("selected-file", selectedFileVar)
	nb.AddNs("navigation",
		eval.BuildNsNamed("edit:navigation").
			AddVars(map[string]vars.Var{
				"binding":     bindingVar,
				"width-ratio": widthRatioVar,
			}).
			AddGoFns(map[string]any{
				"start": func() {
					w, err := modes.NewNavigation(app, modes.NavigationSpec{
						Bindings: bindings,
						Cursor:   modes.NewOSNavigationCursor(ev.Chdir),
						WidthRatio: func() [3]int {
							return convertNavWidthRatio(widthRatioVar.Get())
						},
						Filter: filterSpec,
					})
					if err != nil {
						app.Notify(modes.ErrorText(err))
					} else {
						startMode(app, w, nil)
					}
				},
				"left":  actOnNavigation(app, modes.Navigation.Ascend),
				"right": actOnNavigation(app, modes.Navigation.Descend),
				"up": actOnNavigation(app,
					func(w modes.Navigation) { w.Select(tk.Prev) }),
				"down": actOnNavigation(app,
					func(w modes.Navigation) { w.Select(tk.Next) }),
				"page-up": actOnNavigation(app,
					func(w modes.Navigation) { w.Select(tk.PrevPage) }),
				"page-down": actOnNavigation(app,
					func(w modes.Navigation) { w.Select(tk.NextPage) }),

				"file-preview-up": actOnNavigation(app,
					func(w modes.Navigation) { w.ScrollPreview(-1) }),
				"file-preview-down": actOnNavigation(app,
					func(w modes.Navigation) { w.ScrollPreview(1) }),

				"insert-selected":          func() { navInsertSelected(app) },
				"insert-selected-and-quit": func() { navInsertSelectedAndQuit(app) },

				"trigger-filter": actOnNavigation(app,
					func(w modes.Navigation) { w.MutateFiltering(neg) }),
				"trigger-shown-hidden": actOnNavigation(app,
					func(w modes.Navigation) { w.MutateShowHidden(neg) }),
			}))
}

func neg(b bool) bool { return !b }

func activeNavigation(app cli.App) (modes.Navigation, bool) {
	w, ok := app.ActiveWidget().(modes.Navigation)
	return w, ok
}

func actOnNavigation(app cli.App, f func(modes.Navigation)) func() {
	return func() {
		if w, ok := activeNavigation(app); ok {
			f(w)
		}
	}
}
