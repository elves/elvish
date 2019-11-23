// Package navigation provides the functionality of navigating the filesystem.
package navigation

import (
	"sort"
	"strings"
	"unicode"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/cli/el/colview"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/el/textview"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
)

type widget struct {
	codeArea codearea.Widget
	colView  colview.Widget
}

func (w widget) Handle(e term.Event) bool { return w.colView.Handle(e) }

func (w widget) Render(width, height int) *ui.Buffer {
	buf := w.codeArea.Render(width, height)
	bufColView := w.colView.Render(width, height-len(buf.Lines))
	buf.Extend(bufColView, false)
	return buf
}

// Config contains the configuration needed for the navigation functionality.
type Config struct {
	// Key binding.
	Binding el.Handler
	// Underlying filesystem.
	Cursor Cursor
}

// Start starts the navigation function.
func Start(app cli.App, cfg Config) {
	cursor := cfg.Cursor
	if cursor == nil {
		cursor = NewOSCursor()
	}

	onLeft := func(w colview.Widget) {
		// Remember the name of the current directory before ascending.
		currentName := ""
		current, err := cursor.Current()
		if err == nil {
			currentName = current.Name()
		}

		err = cursor.Ascend()
		if err != nil {
			app.Notify(err.Error())
		} else {
			updateState(w, cursor, currentName)
		}
	}
	onRight := func(w colview.Widget) {
		currentCol, ok := w.CopyState().Columns[1].(listbox.Widget)
		if !ok {
			return
		}
		state := currentCol.CopyState()
		if state.Items.Len() == 0 {
			return
		}
		selected := state.Items.(fileItems)[state.Selected]
		if !selected.Mode().IsDir() {
			// Check if the file is a symlink to a directory.
			mode, err := selected.DeepMode()
			if err != nil {
				app.Notify(err.Error())
				return
			}
			if !mode.IsDir() {
				return
			}
		}
		err := cursor.Descend(selected.Name())
		if err != nil {
			app.Notify(err.Error())
		} else {
			updateState(w, cursor, "")
		}
	}
	w := widget{
		codeArea: codearea.New(codearea.Spec{
			Prompt: layout.ModePrompt(" NAVIGATING ", true),
		}),
		colView: colview.New(colview.Spec{
			OverlayHandler: cfg.Binding,
			Weights:        func(n int) []int { return []int{1, 3, 4} },
			OnLeft:         onLeft,
			OnRight:        onRight,
		}),
	}
	updateState(w.colView, cursor, "")
	app.MutateState(func(s *cli.State) { s.Addon = w })
	app.Redraw()
}

// SelectedName returns the currently selected name in the navigation addon. It
// returns an empty string if the navigation addon is not active.
func SelectedName(app cli.App) string {
	w, ok := app.CopyState().Addon.(widget)
	if !ok {
		return ""
	}
	state := w.colView.CopyState().Columns[1].(listbox.Widget).CopyState()
	return state.Items.(fileItems)[state.Selected].Name()
}

func updateState(w colview.Widget, cursor Cursor, selectName string) {
	var parentCol, currentCol el.Widget

	w.MutateState(func(s *colview.State) {
		*s = colview.State{
			Columns: []el.Widget{
				layout.Empty{}, layout.Empty{}, layout.Empty{}},
			FocusColumn: 1,
		}
	})

	parent, err := cursor.Parent()
	if err == nil {
		parentCol = makeWidget(parent)
	} else {
		parentCol = makeErrWidget(err)
	}

	current, err := cursor.Current()
	if err == nil {
		currentCol = makeWidgetWithOnSelect(
			current,
			func(it listbox.Items, i int) {
				previewCol := makeWidget(it.(fileItems)[i])
				w.MutateState(func(s *colview.State) {
					s.Columns[2] = previewCol
				})
			})
		tryToSelectName(parentCol, current.Name())
		if selectName != "" {
			tryToSelectName(currentCol, selectName)
		}
	} else {
		currentCol = makeErrWidget(err)
		tryToSelectNothing(parentCol)
	}

	w.MutateState(func(s *colview.State) {
		s.Columns[0] = parentCol
		s.Columns[1] = currentCol
	})
}

// Selects nothing if the widget is a listbox.
func tryToSelectNothing(w el.Widget) {
	list, ok := w.(listbox.Widget)
	if !ok {
		return
	}
	list.Select(func(listbox.State) int { return -1 })
}

// Selects the item with the given name, if the widget is a listbox with
// fileItems and has such an item.
func tryToSelectName(w el.Widget, name string) {
	list, ok := w.(listbox.Widget)
	if !ok {
		// Do nothing
		return
	}
	list.Select(func(state listbox.State) int {
		items, ok := state.Items.(fileItems)
		if !ok {
			return 0
		}
		for i, file := range items {
			if file.Name() == name {
				return i
			}
		}
		return 0
	})
}

func makeWidget(f File) el.Widget {
	return makeWidgetWithOnSelect(f, nil)
}

func makeWidgetWithOnSelect(f File, onSelect func(listbox.Items, int)) el.Widget {
	files, content, err := f.Read()
	if err != nil {
		return makeErrWidget(err)
	}

	if files != nil {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})
		return listbox.New(listbox.Spec{
			Padding: 1, ExtendStyle: true, OnSelect: onSelect,
			State: listbox.State{Items: fileItems(files)},
		})
	}

	lines := strings.Split(sanitize(string(content)), "\n")
	return textview.New(textview.Spec{
		State:      textview.State{Lines: lines},
		Scrollable: true,
	})
}

func makeErrWidget(err error) el.Widget {
	return layout.Label{Content: styled.MakeText(err.Error(), "red")}
}

type fileItems []File

func (it fileItems) Show(i int) styled.Text {
	// TODO: Support lsColors
	if it[i].Mode().IsDir() {
		return styled.MakeText(it[i].Name(), "blue")
	}
	return styled.Plain(it[i].Name())
}

func (it fileItems) Len() int { return len(it) }

func sanitize(content string) string {
	// Remove unprintable characters, and replace tabs with 4 spaces.
	var sb strings.Builder
	for _, r := range content {
		if r == '\t' {
			sb.WriteString("    ")
		} else if r == '\n' || unicode.IsGraphic(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Select changes the selection if the navigation addon is currently active.
func Select(app cli.App, f func(listbox.State) int) {
	actOnWidget(app, func(w widget) {
		if listBox, ok := w.colView.CopyState().Columns[1].(listbox.Widget); ok {
			listBox.Select(f)
			app.Redraw()
		}
	})
}

// ScrollPreview scrolls the preview if the navigation addon is currently
// active.
func ScrollPreview(app cli.App, delta int) {
	actOnWidget(app, func(w widget) {
		if textView, ok := w.colView.CopyState().Columns[2].(textview.Widget); ok {
			textView.ScrollBy(delta)
			app.Redraw()
		}
	})
}

// Ascend ascends in the navigation addon if it is active.
func Ascend(app cli.App) {
	actOnWidget(app, func(w widget) {
		w.colView.Left()
		app.Redraw()
	})
}

// Descend descends in the navigation addon if it is active.
func Descend(app cli.App) {
	actOnWidget(app, func(w widget) {
		w.colView.Right()
		app.Redraw()
	})
}

func actOnWidget(app cli.App, f func(widget)) {
	w, ok := app.CopyState().Addon.(widget)
	if ok {
		f(w)
	}
}
