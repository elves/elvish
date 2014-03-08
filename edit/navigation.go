package edit

import (
	"errors"
	"os"
	"sort"
)

var (
	errorEmptyCwd      = errors.New("current directory is empty")
	errorNoCwdInParent = errors.New("could not find current directory in ..")
)

type navColumn struct {
	names    []string
	selected int
}

func newNavColumn(names []string) *navColumn {
	nc := &navColumn{names: names}
	nc.resetSelected()
	return nc
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

func readdirnames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func (n *navigation) maintainSelected(name string) {
	i := sort.SearchStrings(n.current.names, name)
	if i == len(n.current.names) {
		i--
	}
	n.current.selected = i
}

// refresh rereads files in current and parent directories and maintains the
// selected file if possible.
func (n *navigation) refresh() error {
	selectedName := n.current.selectedName()

	// n.current
	names, err := readdirnames(".")
	if err != nil {
		return err
	}
	n.current = newNavColumn(names)

	if selectedName != "" {
		// Maintain n.current.selected. The same file, if still present, is
		// selected. Otherwise a file near it is selected.
		// XXX(xiaq): This would break when we support alternative ordering.
		n.maintainSelected(selectedName)
	}

	// n.parent
	names, err = readdirnames("..")
	if err != nil {
		return err
	}
	n.parent = newNavColumn(names)

	cwd, err := os.Stat(".")
	if err != nil {
		return err
	}
	n.parent.selected = -1
	for i, name := range n.parent.names {
		d, _ := os.Lstat("../" + name)
		if os.SameFile(d, cwd) {
			n.parent.selected = i
			break
		}
	}
	if n.parent.selected == -1 {
		return errorNoCwdInParent
	}

	// n.dirPreview
	if n.current.selected != -1 {
		name := n.current.selectedName()
		fi, err := os.Stat(name)
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			names, err = readdirnames(name)
			if err != nil {
				return err
			}
			n.dirPreview = newNavColumn(names)
		} else {
			// TODO(xiaq): Support regular file preview in navigation mode
			n.dirPreview = nil
		}
	} else {
		n.dirPreview = nil
	}

	return nil
}

// ascend changes current directory to the parent.
// TODO(xiaq): navigation.{ascend descend} bypasses the cd builtin. This can be
// problematic if cd acquires more functionality (e.g. trigger a hook).
func (n *navigation) ascend() error {
	name := n.parent.names[n.parent.selected]
	err := os.Chdir("..")
	if err != nil {
		return err
	}
	err = n.refresh()
	if err != nil {
		return err
	}
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
	return n.refresh()
}

// prev selects the previous file.
func (n *navigation) prev() error {
	if n.current.selected > 0 {
		n.current.selected--
	}
	return n.refresh()
}

// next selects the next file.
func (n *navigation) next() error {
	if n.current.selected != -1 && n.current.selected < len(n.current.names)-1 {
		n.current.selected++
	}
	return n.refresh()
}
