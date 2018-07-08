package newedit

import (
	"os"

	"github.com/elves/elvish/newedit/core"
)

type Editor interface {
	ReadCode() (string, error)
}

func NewEditor(in, out *os.File) Editor {
	ed := core.NewEditor(core.NewTTY(in, out))
	return ed
}
