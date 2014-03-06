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

type navigation struct {
	current, parent navColumn
}

func newNavigation() *navigation {
	n := &navigation{current: navColumn{selected: -1}}
	n.refresh()
	n.resetSelected()
	return n
}

func (n *navigation) resetSelected() {
	if len(n.current.names) > 0 {
		n.current.selected = 0
	} else {
		n.current.selected = -1
	}
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
	selectedName := ""
	if n.current.selected != -1 {
		selectedName = n.current.names[n.current.selected]
	}

	var err error
	n.current.names, err = readdirnames(".")
	if err != nil {
		return err
	}
	n.resetSelected()
	if selectedName != "" {
		// Maintain n.current.selected. The same file, if still present, is selected.
		// Otherwise a file near it is selected.
		// XXX(xiaq): This would break when we support alternative ordering.
		n.maintainSelected(selectedName)
	}

	n.parent.names, err = readdirnames("..")
	if err != nil {
		return err
	}
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
func (n *navigation) prev() {
	if n.current.selected > 0 {
		n.current.selected--
	}
}

// next selects the next file.
func (n *navigation) next() {
	if n.current.selected != -1 && n.current.selected < len(n.current.names)-1 {
		n.current.selected++
	}
}
