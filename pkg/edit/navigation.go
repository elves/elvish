package edit

import (
	"strings"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/mode/navigation"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
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
	fname := navigation.SelectedName(app)
	if fname == "" {
		// User pressed Alt-Enter or Enter in an empty directory with nothing
		// selected; don't do anything.
		return
	}

	app.CodeArea().MutateState(func(s *tk.CodeAreaState) {
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

//elvdoc:var navigation:width-ratio
//
// A list of 3 integers, used for specifying the width ratio of the 3 columns in
// navigation mode.

func convertNavWidthRatio(v interface{}) [3]int {
	var (
		numbers []int
		hasErr  bool
	)
	vals.Iterate(v, func(elem interface{}) bool {
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
	bindingVar := newBindingVar(emptyBindingMap)
	binding := newMapBinding(ed, ev, bindingVar)
	widthRatioVar := newListVar(vals.MakeList(1.0, 3.0, 4.0))

	selectedFileVar := vars.FromGet("edit:selected-file", func() interface{} {
		name := navigation.SelectedName(ed.app)
		if name == "" {
			return nil
		}
		return name
	})

	app := ed.app
	nb.Add("selected-file", selectedFileVar)
	nb.AddNs("navigation",
		eval.NsBuilder{
			"binding":     bindingVar,
			"width-ratio": widthRatioVar,
		}.AddGoFns("<edit:navigation>", map[string]interface{}{
			"start": func() {
				navigation.Start(app, navigation.Config{
					Binding: binding,
					WidthRatio: func() [3]int {
						return convertNavWidthRatio(widthRatioVar.Get())
					},
				})
			},
			"left":      func() { navigation.Ascend(app) },
			"right":     func() { navigation.Descend(app) },
			"up":        func() { navigation.Select(app, tk.Prev) },
			"down":      func() { navigation.Select(app, tk.Next) },
			"page-up":   func() { navigation.Select(app, tk.PrevPage) },
			"page-down": func() { navigation.Select(app, tk.NextPage) },

			"file-preview-up":   func() { navigation.ScrollPreview(app, -1) },
			"file-preview-down": func() { navigation.ScrollPreview(app, 1) },

			"insert-selected":          func() { navInsertSelected(app) },
			"insert-selected-and-quit": func() { navInsertSelectedAndQuit(app) },

			"trigger-filter":       func() { navToggleFilter(app) },
			"trigger-shown-hidden": func() { navToggleShowHidden(app) },
		}).Ns())
}
