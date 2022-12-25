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
	"src.elv.sh/pkg/ui"
)

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

func navInsertSelectedAndQuit(app cli.App) {
	navInsertSelected(app)
	closeMode(app)
}

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
	// TODO: Rename to $edit:navigation:selected-file after deprecation
	nb.AddVar("selected-file", selectedFileVar)
	ns := eval.BuildNsNamed("edit:navigation").
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
					CodeAreaRPrompt: func() ui.Text {
						return bindingTips(ed.ns, "navigation:binding",
							bindingTip("hidden", "navigation:trigger-shown-hidden"),
							bindingTip("filter", "navigation:trigger-filter"))
					},
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
			// TODO: Rename to trigger-show-hidden after deprecation
			"trigger-shown-hidden": actOnNavigation(app,
				func(w modes.Navigation) { w.MutateShowHidden(neg) }),
		}).Ns()
	nb.AddNs("navigation", ns)
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
