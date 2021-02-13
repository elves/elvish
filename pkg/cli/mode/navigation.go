// Package navigation provides the functionality of navigating the filesystem.
package mode

import (
	"sort"
	"strings"
	"sync"
	"unicode"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

type Navigation interface {
	tk.Widget
	// SelectedName returns the currently selected name. It returns an empty
	// string if there is no selected name, which can happen if the current
	// directory is empty.
	SelectedName() string
	// Select changes the selection.
	Select(f func(tk.ListBoxState) int)
	// ScrollPreview scrolls the preview.
	ScrollPreview(delta int)
	// Ascend ascends to the parent directory.
	Ascend()
	// Descend descends into the currently selected child directory.
	Descend()
	// MutateFiltering changes the filtering status.
	MutateFiltering(f func(bool) bool)
	// MutateShowHidden changes whether hidden files - files whose names start
	// with ".", should be shown.
	MutateShowHidden(f func(bool) bool)
}

// NavigationSpec specifieis the configuration for the navigation mode.
type NavigationSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// Underlying filesystem.
	Cursor NavigationCursor
	// A function that returns the relative weights of the widths of the 3
	// columns. If unspecified, the ratio is 1:3:4.
	WidthRatio func() [3]int
}

type navigationState struct {
	Filtering  bool
	ShowHidden bool
}

type navigation struct {
	NavigationSpec
	app        cli.App
	codeArea   tk.CodeArea
	colView    tk.ColView
	lastFilter string
	stateMutex sync.RWMutex
	state      navigationState
}

func (w *navigation) MutateState(f func(*navigationState)) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	f(&w.state)
}

func (w *navigation) CopyState() navigationState {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()
	return w.state
}

func (w *navigation) Handle(event term.Event) bool {
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
		}
		return false
	}
	return w.app.CodeArea().Handle(event)
}

func (w *navigation) Render(width, height int) *term.Buffer {
	buf := w.codeArea.Render(width, height)
	bufColView := w.colView.Render(width, height-len(buf.Lines))
	buf.Extend(bufColView, false)
	return buf
}

func (w *navigation) Focus() bool {
	return w.CopyState().Filtering
}

func (w *navigation) ascend() {
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
		w.codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.Buffer = tk.CodeBuffer{}
		})
		updateState(w, currentName)
	}
}

func (w *navigation) descend() {
	currentCol, ok := w.colView.CopyState().Columns[1].(tk.ListBox)
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
		w.codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.Buffer = tk.CodeBuffer{}
		})
		updateState(w, "")
	}
}

// NewNavigation creates a new navigation mode.
func NewNavigation(app cli.App, spec NavigationSpec) Navigation {
	if spec.Cursor == nil {
		spec.Cursor = NewOSNavigationCursor()
	}
	if spec.WidthRatio == nil {
		spec.WidthRatio = func() [3]int { return [3]int{1, 3, 4} }
	}

	var w *navigation
	w = &navigation{
		NavigationSpec: spec,
		app:            app,
		codeArea: tk.NewCodeArea(tk.CodeAreaSpec{
			Prompt: func() ui.Text {
				if w.CopyState().ShowHidden {
					return ModeLine(" NAVIGATING (show hidden) ", true)
				}
				return ModeLine(" NAVIGATING ", true)
			},
		}),
		colView: tk.NewColView(tk.ColViewSpec{
			Bindings: spec.Bindings,
			Weights: func(int) []int {
				a := spec.WidthRatio()
				return a[:]
			},
			OnLeft:  func(tk.ColView) { w.ascend() },
			OnRight: func(tk.ColView) { w.descend() },
		}),
	}
	updateState(w, "")
	return w
}

func (w *navigation) SelectedName() string {
	col, ok := w.colView.CopyState().Columns[1].(tk.ListBox)
	if !ok {
		return ""
	}
	state := col.CopyState()
	if 0 <= state.Selected && state.Selected < state.Items.Len() {
		return state.Items.(fileItems)[state.Selected].Name()
	}
	return ""
}

func updateState(w *navigation, selectName string) {
	colView := w.colView
	cursor := w.Cursor
	filter := w.lastFilter
	showHidden := w.CopyState().ShowHidden

	var parentCol, currentCol tk.Widget

	colView.MutateState(func(s *tk.ColViewState) {
		*s = tk.ColViewState{
			Columns: []tk.Widget{
				tk.Empty{}, tk.Empty{}, tk.Empty{}},
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
			func(it tk.Items, i int) {
				previewCol := makeCol(it.(fileItems)[i], showHidden)
				colView.MutateState(func(s *tk.ColViewState) {
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

	colView.MutateState(func(s *tk.ColViewState) {
		s.Columns[0] = parentCol
		s.Columns[1] = currentCol
	})
}

// Selects nothing if the widget is a listbox.
func tryToSelectNothing(w tk.Widget) {
	list, ok := w.(tk.ListBox)
	if !ok {
		return
	}
	list.Select(func(tk.ListBoxState) int { return -1 })
}

// Selects the item with the given name, if the widget is a listbox with
// fileItems and has such an item.
func tryToSelectName(w tk.Widget, name string) {
	list, ok := w.(tk.ListBox)
	if !ok {
		// Do nothing
		return
	}
	list.Select(func(state tk.ListBoxState) int {
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

func makeCol(f NavigationFile, showHidden bool) tk.Widget {
	return makeColInner(f, "", showHidden, nil)
}

func makeColInner(f NavigationFile, filter string, showHidden bool, onSelect func(tk.Items, int)) tk.Widget {
	files, content, err := f.Read()
	if err != nil {
		return makeErrCol(err)
	}

	if files != nil {
		if filter != "" || !showHidden {
			var filtered []NavigationFile
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
		return tk.NewListBox(tk.ListBoxSpec{
			Padding: 1, ExtendStyle: true, OnSelect: onSelect,
			State: tk.ListBoxState{Items: fileItems(files)},
		})
	}

	lines := strings.Split(sanitize(string(content)), "\n")
	return tk.NewTextView(tk.TextViewSpec{
		State:      tk.TextViewState{Lines: lines},
		Scrollable: true,
	})
}

func makeErrCol(err error) tk.Widget {
	return tk.Label{Content: ui.T(err.Error(), ui.FgRed)}
}

type fileItems []NavigationFile

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

func (w *navigation) Select(f func(tk.ListBoxState) int) {
	if listBox, ok := w.colView.CopyState().Columns[1].(tk.ListBox); ok {
		listBox.Select(f)
	}
}

func (w *navigation) ScrollPreview(delta int) {
	if textView, ok := w.colView.CopyState().Columns[2].(tk.TextView); ok {
		textView.ScrollBy(delta)
	}
}

func (w *navigation) Ascend() {
	w.colView.Left()
}

func (w *navigation) Descend() {
	w.colView.Right()
}

func (w *navigation) MutateFiltering(f func(bool) bool) {
	w.MutateState(func(s *navigationState) { s.Filtering = f(s.Filtering) })
}

func (w *navigation) MutateShowHidden(f func(bool) bool) {
	w.MutateState(func(s *navigationState) { s.ShowHidden = f(s.ShowHidden) })
	updateState(w, w.SelectedName())
}
