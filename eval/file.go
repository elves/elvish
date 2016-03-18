package eval

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
)

type File struct {
	inner *os.File
}

var _ Value = File{}

func (File) Kind() string {
	return "file"
}

func (f File) Repr(int) string {
	return fmt.Sprintf("<file{%s %p}>", parse.Quote(f.inner.Name()), f.inner)
}
