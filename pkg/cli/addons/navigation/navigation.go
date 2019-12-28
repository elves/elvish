// Package navigation provides the functionality of navigating the filesystem.
package navigation

import (
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

// Config contains the configuration needed for the navigation functionality.
type Config struct {
	// Key binding.
	Binding cli.Handler
	// Underlying filesystem.
	Cursor Cursor
	// A function that returns the relative weights of the widths of the 3
	// columns. If unspecified, the ratio is 1:3:4.
	WidthRatio func() [3]int
}

type state struct {
	Filtering  bool
	ShowHidden bool
}

type widget struct {
	Config
	app        cli.App
	codeArea   cli.CodeArea
	colView    cli.ColView
	lastFilter string
	stateMutex sync.RWMutex
	state      state
}

func (w *widget) MutateState(f func(*state)) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	f(&w.state)
}

func (w *widget) CopyState() state {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()
	return w.state
}

func (w *widget) Handle(event term.Event) bool {
	if w.colView.Handle(event) {
		return true
	}
	if w.CopyState().Filtering {
		if w.codeArea.Handle(event) {
			filter := w.codeArea.CopyState().Buffer.Content
			if filter != w.lastFilter {
				w.lastFilter = filter
				updateState(w, "")
			}
			return true
		} else {
			return false
		}
	} else {
		return w.app.CodeArea().Handle(event)
	}
}

func (w *widget) Render(width, height int) *term.Buffer {
	buf := w.codeArea.Render(width, height)
	bufColView := w.colView.Render(width, height-len(buf.Lines))
	buf.Extend(bufColView, false)
	return buf
}

func (w *widget) Focus() bool {
	return w.CopyState().Filtering
}

func (w *widget) ascend() {
	// Remember the name of the current directory before ascending.
	currentName := ""
	current, err := w.Cursor.Current()
	if err == nil {
		currentName = current.Name()
	}

	err = w.Cursor.Ascend()
	if err != nil {
		w.app.Notify(err.Error())
	} else {
		w.codeArea.MutateState(func(s *cli.CodeAreaState) {
			s.Buffer = cli.CodeBuffer{}
		})
		updateState(w, currentName)
	}
}

func (w *widget) descend() {
	currentCol, ok := w.colView.CopyState().Columns[1].(cli.ListBox)
	if !ok {
		return
	}
	state := currentCol.CopyState()
	if state.Items.Len() == 0 {
		return
	}
	selected := state.Items.(fileItems)[state.Selected]
	if !selected.IsDirDeep() {
		return
	}
	err := w.Cursor.Descend(selected.Name())
	if err != nil {
		w.app.Notify(err.Error())
	} else {
		w.codeArea.MutateState(func(s *cli.CodeAreaState) {
			s.Buffer = cli.CodeBuffer{}
		})
		updateState(w, "")
	}
}

// Start starts the navigation function.
func Start(app cli.App, cfg Config) {
	if cfg.Cursor == nil {
		cfg.Cursor = NewOSCursor()
	}
	if cfg.WidthRatio == nil {
		cfg.WidthRatio = func() [3]int { return [3]int{1, 3, 4} }
	}

	var w *widget
	w = &widget{
		Config: cfg,
		app:    app,
		codeArea: cli.NewCodeArea(cli.CodeAreaSpec{
			Prompt: func() ui.Text {
				if w.CopyState().ShowHidden {
					return cli.ModeLine(" NAVIGATING (show hidden) ", true)
				} else {
					return cli.ModeLine(" NAVIGATING ", true)
				}
			},
		}),
		colView: cli.NewColView(cli.ColViewSpec{
			OverlayHandler: cfg.Binding,
			Weights: func(int) []int {
				a := cfg.WidthRatio()
				return a[:]
			},
			OnLeft:  func(cli.ColView) { w.ascend() },
			OnRight: func(cli.ColView) { w.descend() },
		}),
	}
	updateState(w, "")
	app.MutateState(func(s *cli.State) { s.Addon = w })
	app.Redraw()
}

// SelectedName returns the currently selected name in the navigation addon. It
// returns an empty string if the navigation addon is not active, or if there is
// no selected name.
func SelectedName(app cli.App) string {
	w, ok := app.CopyState().Addon.(*widget)
	if !ok {
		return ""
	}
	col, ok := w.colView.CopyState().Columns[1].(cli.ListBox)
	if !ok {
		return ""
	}
	state := col.CopyState()
	if 0 <= state.Selected && state.Selected < state.Items.Len() {
		return state.Items.(fileItems)[state.Selected].Name()
	}
	return ""
}

func updateState(w *widget, selectName string) {
	colView := w.colView
	cursor := w.Cursor
	filter := w.lastFilter
	showHidden := w.CopyState().ShowHidden

	var parentCol, currentCol cli.Widget

	colView.MutateState(func(s *cli.ColViewState) {
		*s = cli.ColViewState{
			Columns: []cli.Widget{
				cli.Empty{}, cli.Empty{}, cli.Empty{}},
			FocusColumn: 1,
		}
	})

	parent, err := cursor.Parent()
	if err == nil {
		parentCol = makeCol(parent, showHidden)
	} else {
		parentCol = makeErrCol(err)
	}

	current, err := cursor.Current()
	if err == nil {
		currentCol = makeColInner(
			current,
			filter,
			showHidden,
			func(it cli.Items, i int) {
				previewCol := makeCol(it.(fileItems)[i], showHidden)
				colView.MutateState(func(s *cli.ColViewState) {
					s.Columns[2] = previewCol
				})
			})
		tryToSelectName(parentCol, current.Name())
		if selectName != "" {
			tryToSelectName(currentCol, selectName)
		}
	} else {
		currentCol = makeErrCol(err)
		tryToSelectNothing(parentCol)
	}

	colView.MutateState(func(s *cli.ColViewState) {
		s.Columns[0] = parentCol
		s.Columns[1] = currentCol
	})
}

// Selects nothing if the widget is a cli.
func tryToSelectNothing(w cli.Widget) {
	list, ok := w.(cli.ListBox)
	if !ok {
		return
	}
	list.Select(func(cli.ListBoxState) int { return -1 })
}

// Selects the item with the given name, if the widget is a listbox with
// fileItems and has such an item.
func tryToSelectName(w cli.Widget, name string) {
	list, ok := w.(cli.ListBox)
	if !ok {
		// Do nothing
		return
	}
	list.Select(func(state cli.ListBoxState) int {
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

func makeCol(f File, showHidden bool) cli.Widget {
	return makeColInner(f, "", showHidden, nil)
}

func makeColInner(f File, filter string, showHidden bool, onSelect func(cli.Items, int)) cli.Widget {
	files, content, err := f.Read()
	if err != nil {
		return makeErrCol(err)
	}

	if files != nil {
		if filter != "" || !showHidden {
			var filtered []File
			for _, file := range files {
				name := file.Name()
				hidden := len(name) > 0 && name[0] == '.'
				if strings.Contains(name, filter) && (showHidden || !hidden) {
					filtered = append(filtered, file)
				}
			}
			files = filtered
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})
		return cli.NewListBox(cli.ListBoxSpec{
			Padding: 1, ExtendStyle: true, OnSelect: onSelect,
			State: cli.ListBoxState{Items: fileItems(files)},
		})
	}

	lines := strings.Split(sanitize(string(content)), "\n")
	return cli.NewTextView(cli.TextViewSpec{
		State:      cli.TextViewState{Lines: lines},
		Scrollable: true,
	})
}

func makeErrCol(err error) cli.Widget {
	return cli.Label{Content: ui.T(err.Error(), ui.FgRed)}
}

type fileItems []File

func (it fileItems) Show(i int) ui.Text {
	return it[i].ShowName()
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
func Select(app cli.App, f func(cli.ListBoxState) int) {
	actOnWidget(app, func(w *widget) {
		if listBox, ok := w.colView.CopyState().Columns[1].(cli.ListBox); ok {
			listBox.Select(f)
			app.Redraw()
		}
	})
}

// ScrollPreview scrolls the preview if the navigation addon is currently
// active.
func ScrollPreview(app cli.App, delta int) {
	actOnWidget(app, func(w *widget) {
		if textView, ok := w.colView.CopyState().Columns[2].(cli.TextView); ok {
			textView.ScrollBy(delta)
			app.Redraw()
		}
	})
}

// Ascend ascends in the navigation addon if it is active.
func Ascend(app cli.App) {
	actOnWidget(app, func(w *widget) {
		w.colView.Left()
		app.Redraw()
	})
}

// Descend descends in the navigation addon if it is active.
func Descend(app cli.App) {
	actOnWidget(app, func(w *widget) {
		w.colView.Right()
		app.Redraw()
	})
}

// MutateFiltering changes the filtering status of the navigation addon if it is
// active.
func MutateFiltering(app cli.App, f func(bool) bool) {
	actOnWidget(app, func(w *widget) {
		w.MutateState(func(s *state) { s.Filtering = f(s.Filtering) })
		app.Redraw()
	})
}

// MutateFiltering changes whether the navigation addon should show file whose
// names start with ".".
func MutateShowHidden(app cli.App, f func(bool) bool) {
	actOnWidget(app, func(w *widget) {
		w.MutateState(func(s *state) { s.ShowHidden = f(s.ShowHidden) })
		updateState(w, SelectedName(app))
		app.Redraw()
	})
}

func actOnWidget(app cli.App, f func(*widget)) {
	w, ok := app.CopyState().Addon.(*widget)
	if ok {
		f(w)
	}
}
