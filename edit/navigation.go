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

type navigation struct {
	filenames       []string
	parentFilenames []string
	selected        int
	selectedParent  int
	error           string
}

func newNavigation() *navigation {
	n := &navigation{selected: -1}
	n.refresh()
	n.resetSelected()
	return n
}

func (n *navigation) resetSelected() {
	if len(n.filenames) > 0 {
		n.selected = 0
	} else {
		n.selected = -1
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
	i := sort.SearchStrings(n.filenames, name)
	if i == len(n.filenames) {
		i--
	}
	n.selected = i
}

// refresh rereads files in current and parent directories and maintains the
// selected file if possible.
func (n *navigation) refresh() error {
	selectedName := ""
	if n.selected != -1 {
		selectedName = n.filenames[n.selected]
	}

	var err error
	n.filenames, err = readdirnames(".")
	if err != nil {
		return err
	}
	n.resetSelected()
	if selectedName != "" {
		// Maintain n.selected. The same file, if still present, is selected.
		// Otherwise a file near it is selected.
		// XXX(xiaq): This would break when we support alternative ordering.
		n.maintainSelected(selectedName)
	}

	n.parentFilenames, err = readdirnames("..")
	if err != nil {
		return err
	}
	cwd, err := os.Stat(".")
	if err != nil {
		return err
	}
	n.selectedParent = -1
	for i, name := range n.parentFilenames {
		d, _ := os.Lstat("../" + name)
		if os.SameFile(d, cwd) {
			n.selectedParent = i
			break
		}
	}
	if n.selectedParent == -1 {
		return errorNoCwdInParent
	}
	return nil
}

// ascend changes current directory to the parent.
// TODO(xiaq): navigation.{ascend descend} bypasses the cd builtin. This can be
// problematic if cd acquires more functionality (e.g. trigger a hook).
func (n *navigation) ascend() error {
	name := n.parentFilenames[n.selectedParent]
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
	if n.selected == -1 {
		return errorEmptyCwd
	}
	name := n.filenames[n.selected]
	err := os.Chdir(name)
	if err != nil {
		return err
	}
	return n.refresh()
}

// prev selects the previous file.
func (n *navigation) prev() {
	if n.selected > 0 {
		n.selected--
	}
}

// next selects the next file.
func (n *navigation) next() {
	if n.selected != -1 && n.selected < len(n.filenames)-1 {
		n.selected++
	}
}
