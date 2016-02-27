package edit

import (
	"errors"
	"os"
	"path"
	"sort"

	"github.com/elves/elvish/parse"
)

// Navigation subsystem.

// Interface.

type navigation struct {
	current    *navColumn
	parent     *navColumn
	dirPreview *navColumn
	showHidden bool
}

func (*navigation) Mode() ModeType {
	return modeNavigation
}

func (*navigation) ModeLine(width int) *buffer {
	return makeModeLine(" NAVIGATING ", width)
}

func startNavigation(ed *Editor) {
	initNavigation(&ed.navigation)
	ed.mode = &ed.navigation
}

func selectNavUp(ed *Editor) {
	ed.navigation.prev()
}

func selectNavDown(ed *Editor) {
	ed.navigation.next()
}

func ascendNav(ed *Editor) {
	ed.navigation.ascend()
}

func descendNav(ed *Editor) {
	ed.navigation.descend()
}

func triggerNavShowHidden(ed *Editor) {
	ed.navigation.showHidden = !ed.navigation.showHidden
	ed.navigation.refresh()
}

func navInsertSelected(ed *Editor) {
	ed.insertAtDot(parse.Quote(ed.navigation.current.selectedName()))
}

func quitNavigation(ed *Editor) {
	ed.mode = &ed.insert
}

func defaultNavigation(ed *Editor) {
	// Use key binding for insert mode without exiting navigation mode.
	if f, ok := keyBindings[modeInsert][ed.lastKey]; ok {
		f.Call(ed)
	} else {
		keyBindings[modeInsert][Default].Call(ed)
	}
}

// Implementation.
// TODO(xiaq): Support file preview in navigation mode
// TODO(xiaq): Remember which file was selected in each directory.

var (
	errorEmptyCwd      = errors.New("current directory is empty")
	errorNoCwdInParent = errors.New("could not find current directory in ..")
)

func initNavigation(n *navigation) {
	*n = navigation{}
	n.refresh()
}

func (n *navigation) maintainSelected(name string) {
	i := sort.SearchStrings(n.current.names, name)
	if i == len(n.current.names) {
		i--
	}
	n.current.selected = i
}

func (n *navigation) refreshCurrent() {
	selectedName := n.current.selectedName()
	names, styles, err := n.readdirnames(".")
	if err != nil {
		n.current = newErrNavColumn(err)
		return
	}
	n.current = newNavColumn(names, styles)
	if selectedName != "" {
		// Maintain n.current.selected. The same file, if still present, is
		// selected. Otherwise a file near it is selected.
		// XXX(xiaq): This would break when we support alternative
		// ordering.
		n.maintainSelected(selectedName)
	}
}

func (n *navigation) refreshParent() {
	wd, err := os.Getwd()
	if err != nil {
		n.parent = newErrNavColumn(err)
		return
	}
	if wd == "/" {
		n.parent = newNavColumn(nil, nil)
	} else {
		names, styles, err := n.readdirnames("..")
		if err != nil {
			n.parent = newErrNavColumn(err)
			return
		}
		n.parent = newNavColumn(names, styles)

		cwd, err := os.Stat(".")
		if err != nil {
			n.parent = newErrNavColumn(err)
			return
		}
		n.parent.selected = -1
		for i, name := range n.parent.names {
			d, _ := os.Lstat("../" + name)
			if os.SameFile(d, cwd) {
				n.parent.selected = i
				break
			}
		}
	}
}

func (n *navigation) refreshDirPreview() {
	if n.current.selected != -1 {
		name := n.current.selectedName()
		fi, err := os.Stat(name)
		if err != nil {
			n.dirPreview = newErrNavColumn(err)
			return
		}
		if fi.Mode().IsDir() {
			names, styles, err := n.readdirnames(name)
			if err != nil {
				n.dirPreview = newErrNavColumn(err)
				return
			}
			n.dirPreview = newNavColumn(names, styles)
		} else {
			// TODO(xiaq): Support regular file preview in navigation mode
			n.dirPreview = nil
		}
	} else {
		n.dirPreview = nil
	}
}

// refresh rereads files in current and parent directories and maintains the
// selected file if possible.
func (n *navigation) refresh() {
	n.refreshCurrent()
	n.refreshParent()
	n.refreshDirPreview()
}

// ascend changes current directory to the parent.
// TODO(xiaq): navigation.{ascend descend} bypasses the cd builtin. This can be
// problematic if cd acquires more functionality (e.g. trigger a hook).
func (n *navigation) ascend() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if wd == "/" {
		return nil
	}

	name := n.parent.names[n.parent.selected]
	err = os.Chdir("..")
	if err != nil {
		return err
	}
	n.refresh()
	n.maintainSelected(name)
	// XXX Refresh dir preview again. We should perhaps not have used refresh
	// above.
	n.refreshDirPreview()
	return nil
}

// descend changes current directory to the selected file, if it is a
// directory.
func (n *navigation) descend() error {
	if n.current.selected == -1 {
		return errorEmptyCwd
	}
	name := n.current.names[n.current.selected]
	err := os.Chdir(name)
	if err != nil {
		return err
	}
	n.refresh()
	n.current.resetSelected()
	n.refreshDirPreview()
	return nil
}

// prev selects the previous file.
func (n *navigation) prev() {
	if n.current.selected > 0 {
		n.current.selected--
	}
	n.refresh()
}

// next selects the next file.
func (n *navigation) next() {
	if n.current.selected != -1 && n.current.selected < len(n.current.names)-1 {
		n.current.selected++
	}
	n.refresh()
}

// navColumn is a column in the navigation layout.
type navColumn struct {
	names    []string
	styles   []string
	selected int
	err      error
}

func newNavColumn(names, styles []string) *navColumn {
	nc := &navColumn{names, styles, 0, nil}
	nc.resetSelected()
	return nc
}

func newErrNavColumn(err error) *navColumn {
	return &navColumn{err: err}
}

func (nc *navColumn) selectedName() string {
	if nc == nil || nc.selected == -1 {
		return ""
	}
	return nc.names[nc.selected]
}

func (nc *navColumn) resetSelected() {
	if nc == nil {
		return
	}
	if len(nc.names) > 0 {
		nc.selected = 0
	} else {
		nc.selected = -1
	}
}

func (n *navigation) readdirnames(dir string) (names, styles []string, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, nil, err
	}
	infos, err := f.Readdir(0)
	if err != nil {
		return nil, nil, err
	}
	for _, info := range infos {
		if n.showHidden || info.Name()[0] != '.' {
			names = append(names, info.Name())
		}
	}
	sort.Strings(names)

	styles = make([]string, len(names))
	for i, name := range names {
		styles[i] = defaultLsColor.getStyle(path.Join(dir, name))
	}
	return names, styles, nil
}

func (nav *navigation) List(width, maxHeight int) *buffer {
	margin := navigationListingColMargin
	var ratioParent, ratioCurrent, ratioPreview int
	if nav.dirPreview != nil {
		ratioParent = 15
		ratioCurrent = 40
		ratioPreview = 45
	} else {
		ratioParent = 15
		ratioCurrent = 75
		// Leave some space at the right side
	}

	w := width - margin*2

	wParent := w * ratioParent / 100
	wCurrent := w * ratioCurrent / 100
	wPreview := w * ratioPreview / 100

	b := renderNavColumn(nav.parent, wParent, maxHeight)

	bCurrent := renderNavColumn(nav.current, wCurrent, maxHeight)
	b.extendHorizontal(bCurrent, wParent, margin)

	if wPreview > 0 {
		bPreview := renderNavColumn(nav.dirPreview, wPreview, maxHeight)
		b.extendHorizontal(bPreview, wParent+wCurrent+margin, margin)
	}

	return b
}

func renderNavColumn(nc *navColumn, w, h int) *buffer {
	b := newBuffer(w)
	low, high := findWindow(len(nc.names), nc.selected, h)
	for i := low; i < high; i++ {
		if i > low {
			b.newline()
		}
		text := nc.names[i]
		style := nc.styles[i]
		if i == nc.selected {
			style += styleForSelected
		}
		if w >= navigationListingMinWidthForPadding {
			padding := navigationListingColPadding
			b.writePadding(padding, style)
			b.writes(ForceWcWidth(text, w-2), style)
			b.writePadding(padding, style)
		} else {
			b.writes(ForceWcWidth(text, w), style)
		}
	}
	return b
}
