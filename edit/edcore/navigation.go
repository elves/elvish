package edcore

import (
	"errors"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/lscolors"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

// Navigation subsystem.

// Interface.

type navigation struct {
	binding eddefs.BindingMap
	chdir   func(string) error
	navigationState
}

type navigationState struct {
	current    *navColumn
	parent     *navColumn
	preview    navPreview
	showHidden bool
	filtering  bool
	filter     string
}

func init() { atEditorInit(initNavigation) }

func initNavigation(ed *editor, ns eval.Ns) {
	n := &navigation{
		binding: emptyBindingMap,
		chdir:   ed.Evaler().Chdir,
	}
	ed.navigation = n

	subns := eval.Ns{
		"binding": vars.FromPtr(&n.binding),
	}
	subns.AddBuiltinFns("edit:navigation:", map[string]interface{}{
		"start":                    func() { n.start(ed) },
		"up":                       n.prev,
		"down":                     n.next,
		"page-up":                  n.pageUp,
		"page-down":                n.pageDown,
		"left":                     n.ascend,
		"right":                    n.descend,
		"file-preview-up":          n.filePreviewUp,
		"file-preview-down":        n.filePreviewDown,
		"trigger-shown-hidden":     n.triggerShowHidden,
		"trigger-filter":           n.triggerFilter,
		"insert-selected":          func() { n.insertSelected(ed) },
		"insert-selected-and-quit": func() { n.insertSelectedAndQuit(ed) },
		"default":                  func() { n.defaultFn(ed) },
	})
	ns.AddNs("navigation", subns)
}

type navPreview interface {
	FullWidth(int) int
	List(int) ui.Renderer
}

func (n *navigation) Teardown() {
	n.navigationState = navigationState{}
}

func (n *navigation) Binding(k ui.Key) eval.Callable {
	return n.binding.GetOrDefault(k)
}

func (n *navigation) ModeLine() ui.Renderer {
	title := " NAVIGATING "
	if n.showHidden {
		title += "(show hidden) "
	}
	return ui.NewModeLineRenderer(title, n.filter)
}

func (n *navigation) CursorOnModeLine() bool {
	return n.filtering
}

func (n *navigation) start(ed *editor) {
	n.refresh()
	ed.SetMode(n)
}

func (n *navigation) pageUp() {
	n.current.pageUp()
	n.refresh()
}

func (n *navigation) pageDown() {
	n.current.pageDown()
	n.refresh()
}

func (n *navigation) filePreviewUp() {
	fp, ok := n.preview.(*navFilePreview)
	if ok {
		if fp.beginLine > 0 {
			fp.beginLine--
		}
	}
}

func (n *navigation) filePreviewDown() {
	fp, ok := n.preview.(*navFilePreview)
	if ok {
		if fp.beginLine < len(fp.lines)-1 {
			fp.beginLine++
		}
	}
}

func (n *navigation) triggerShowHidden() {
	n.showHidden = !n.showHidden
	n.refresh()
}

func (n *navigation) triggerFilter() {
	n.filtering = !n.filtering
}

func (n *navigation) insertSelected(ed *editor) {
	ed.InsertAtDot(parse.Quote(n.current.selectedName()) + " ")
}

func (n *navigation) insertSelectedAndQuit(ed *editor) {
	ed.InsertAtDot(parse.Quote(n.current.selectedName()) + " ")
	ed.SetModeInsert()
}

func (n *navigation) defaultFn(ed *editor) {
	// Use key binding for insert mode without exiting nigation mode.
	k := ed.lastKey
	if n.filtering && likeChar(k) {
		n.filter += k.String()
		n.refreshCurrent()
		n.refreshDirPreview()
	} else if n.filtering && k == (ui.Key{ui.Backspace, 0}) {
		_, size := utf8.DecodeLastRuneInString(n.filter)
		if size > 0 {
			n.filter = n.filter[:len(n.filter)-size]
			n.refreshCurrent()
			n.refreshDirPreview()
		}
	} else {
		fn := ed.insert.binding.GetOrDefault(k)
		if fn == nil {
			ed.Notify("key %s unbound and no default binding", k)
		} else {
			ed.CallFn(fn)
		}
	}
}

// Implementation.
// TODO(xiaq): Remember which file was selected in each directory.

var errorEmptyCwd = errors.New("current directory is empty")

func (n *navigation) maintainSelected(name string) {
	n.current.selected = 0
	for i, s := range n.current.candidates {
		if s.Text > name {
			break
		}
		n.current.selected = i
	}
}

func (n *navigation) refreshCurrent() {
	selectedName := n.current.selectedName()
	all, err := n.loaddir(".")
	if err != nil {
		n.current = newErrNavColumn(err)
		return
	}
	// Try to select the old selected file.
	// XXX(xiaq): This would break when we support alternative ordering.
	n.current = newNavColumn(all, func(i int) bool {
		return i == 0 || all[i].Text <= selectedName
	})
	n.current.changeFilter(n.filter)
	n.maintainSelected(selectedName)
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
		all, err := n.loaddir("..")
		if err != nil {
			n.parent = newErrNavColumn(err)
			return
		}
		cwd, err := os.Stat(".")
		if err != nil {
			n.parent = newErrNavColumn(err)
			return
		}
		n.parent = newNavColumn(all, func(i int) bool {
			d, _ := os.Lstat("../" + all[i].Text)
			return os.SameFile(d, cwd)
		})
	}
}

func (n *navigation) refreshDirPreview() {
	if n.current.selected != -1 {
		name := n.current.selectedName()
		fi, err := os.Stat(name)
		if err != nil {
			n.preview = newErrNavColumn(err)
			return
		}
		if fi.Mode().IsDir() {
			all, err := n.loaddir(name)
			if err != nil {
				n.preview = newErrNavColumn(err)
				return
			}
			n.preview = newNavColumn(all, func(int) bool { return false })
		} else {
			n.preview = makeNavFilePreview(name)
		}
	} else {
		n.preview = nil
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

	name := n.parent.selectedName()
	err = os.Chdir("..")
	if err != nil {
		return err
	}
	n.filter = ""
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
	name := n.current.selectedName()
	err := n.chdir(name)
	if err != nil {
		return err
	}
	n.filter = ""
	n.current.selected = -1
	n.refresh()
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
	if n.current.selected != -1 && n.current.selected < len(n.current.candidates)-1 {
		n.current.selected++
	}
	n.refresh()
}

func (n *navigation) loaddir(dir string) ([]ui.Styled, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	var all []ui.Styled
	lsColor := lscolors.GetColorist()
	for _, name := range names {
		if n.showHidden || name[0] != '.' {
			all = append(all, ui.Styled{name,
				ui.StylesFromString(lsColor.GetStyle(path.Join(dir, name)))})
		}
	}

	return all, nil
}

func (n *navigation) List(maxHeight int) ui.Renderer {
	return makeNavRenderer(
		maxHeight,
		n.parent.FullWidth(maxHeight),
		n.current.FullWidth(maxHeight),
		n.preview.FullWidth(maxHeight),
		n.parent.List(maxHeight),
		n.current.List(maxHeight),
		n.preview.List(maxHeight),
	)
}

// navColumn is a column in the navigation layout.
type navColumn struct {
	listingMode
	all        []ui.Styled
	candidates []ui.Styled
	// selected int
	err error
}

func newNavColumn(all []ui.Styled, sel func(int) bool) *navColumn {
	nc := &navColumn{all: all, candidates: all}
	nc.provider = nc
	nc.selected = -1
	for i := range all {
		if sel(i) {
			nc.selected = i
		}
	}
	return nc
}

func newErrNavColumn(err error) *navColumn {
	nc := &navColumn{err: err}
	nc.provider = nc
	return nc
}

func (nc *navColumn) Placeholder() string {
	if nc.err != nil {
		return nc.err.Error()
	}
	return ""
}

func (nc *navColumn) Len() int {
	return len(nc.candidates)
}

func (nc *navColumn) Show(i int) (string, ui.Styled) {
	cand := nc.candidates[i]
	return "", ui.Styled{" " + cand.Text + " ", cand.Styles}
}

func (nc *navColumn) Filter(filter string) int {
	nc.candidates = nc.candidates[:0]
	for _, s := range nc.all {
		if strings.Contains(s.Text, filter) {
			nc.candidates = append(nc.candidates, s)
		}
	}
	return 0
}

func (nc *navColumn) FullWidth(h int) int {
	if nc == nil {
		return 0
	}
	if nc.err != nil {
		return util.Wcswidth(nc.err.Error())
	}
	maxw := 0
	for _, s := range nc.candidates {
		maxw = max(maxw, util.Wcswidth(s.Text)+2)
	}
	if len(nc.candidates) > h {
		maxw++
	}
	return maxw
}

func (nc *navColumn) Accept(i int, ed eddefs.Editor) {
	// TODO
}

func (nc *navColumn) ModeTitle(i int) string {
	// Not used
	return ""
}

func (nc *navColumn) selectedName() string {
	if nc == nil || nc.selected == -1 || nc.selected >= len(nc.candidates) {
		return ""
	}
	return nc.candidates[nc.selected].Text
}
