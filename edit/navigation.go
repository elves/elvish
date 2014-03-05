package edit

import (
	"errors"
	"os"
)

var (
	errorEmptyCwd = errors.New("current directory is empty")
)

type navigation struct {
	filenames       []string
	parentFilenames []string
	selected        int
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
	return names, nil
}

// refresh rereads files in current and parent directories and maintains the
// selected file if possible.
func (n *navigation) refresh() error {
	/*
		selectedName := ""
		if n.selected != -1 {
			selectedName = n.filenames[n.selected]
		}
	*/

	var err error
	n.filenames, err = readdirnames(".")
	if err != nil {
		return err
	}
	if n.selected != -1 {
		// TODO(xiaq): Maintain selected
		n.resetSelected()
	}

	n.parentFilenames, err = readdirnames("..")
	if err != nil {
		return err
	}
	return nil
}

// ascend changes current directory to the parent.
// TODO(xiaq): navigation.{ascend descend} bypasses the cd builtin. This can be
// problematic if cd acquires more functionality (e.g. trigger a hook).
func (n *navigation) ascend() error {
	err := os.Chdir("..")
	if err != nil {
		return err
	}
	return n.refresh()
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
	if n.selected != -1 && n.selected < len(n.filenames) {
		n.selected++
	}
}
