package edit

import (
	"errors"
	"os"
	"path"
	"sort"
)

var (
	errorEmptyCwd      = errors.New("current directory is empty")
	errorNoCwdInParent = errors.New("could not find current directory in ..")
)

type navColumn struct {
	names    []string
	attrs    []string
	selected int
	err      error
}

func newNavColumn(names, attrs []string) *navColumn {
	nc := &navColumn{names, attrs, 0, nil}
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

// TODO(xiaq): Handle pwd = / correctly in navigation mode
// TODO(xiaq): Support file preview in navigation mode
type navigation struct {
	current, parent, dirPreview *navColumn
}

func newNavigation() *navigation {
	n := &navigation{}
	n.refresh()
	return n
}

func readdirnames(dir string) (names, attrs []string, err error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, nil, err
	}
	names, err = f.Readdirnames(0)
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(names)
	attrs = make([]string, len(names))
	for i, name := range names {
		attrs[i] = defaultLsColor.determineAttr(path.Join(dir, name))
	}
	return names, attrs, nil
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

	// n.current
	names, attrs, err := readdirnames(".")
	if err != nil {
		n.current = newErrNavColumn(err)
		return
	}
	n.current = newNavColumn(names, attrs)
	if selectedName != "" {
		// Maintain n.current.selected. The same file, if still present, is
		// selected. Otherwise a file near it is selected.
		// XXX(xiaq): This would break when we support alternative
		// ordering.
		n.maintainSelected(selectedName)
	}
}

func (n *navigation) refreshParent() {
	// n.parent
	wd, err := os.Getwd()
	if err != nil {
		n.parent = newErrNavColumn(err)
		return
	}
	if wd == "/" {
		n.parent = newNavColumn(nil, nil)
	} else {
		names, attrs, err := readdirnames("..")
		if err != nil {
			n.parent = newErrNavColumn(err)
			return
		}
		n.parent = newNavColumn(names, attrs)

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
		/*
			if n.parent.selected == -1 {
				return errorNoCwdInParent
			}
		*/
	}
}

func (n *navigation) refreshDirPreview() {
	// n.dirPreview
	if n.current.selected != -1 {
		name := n.current.selectedName()
		fi, err := os.Stat(name)
		if err != nil {
			n.dirPreview = newErrNavColumn(err)
			return
		}
		if fi.Mode().IsDir() {
			names, attrs, err := readdirnames(name)
			if err != nil {
				n.dirPreview = newErrNavColumn(err)
				return
			}
			n.dirPreview = newNavColumn(names, attrs)
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
