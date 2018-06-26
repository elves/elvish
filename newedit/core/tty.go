package core

import (
	"fmt"
	"os"

	"github.com/elves/elvish/edit/tty"
	"github.com/elves/elvish/sys"
)

type TTY interface {
	Setup() (restore func(), err error)
	Size() (h, w int)
}

type realTTY struct {
	in, out *os.File
}

func newTTY(in, out *os.File) TTY {
	return &realTTY{in, out}
}

func (t *realTTY) Setup() (func(), error) {
	restore, err := tty.Setup(t.in, t.out)
	return func() {
		err := restore()
		if err != nil {
			fmt.Println(t.out, "failed to restore terminal properties:", err)
		}
	}, err
}

func (t *realTTY) Size() (h, w int) {
	return sys.GetWinsize(t.out)
}
