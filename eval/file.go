package eval

import (
	"fmt"
	"os"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

type File struct {
	inner *os.File
}

var _ types.Value = File{}

func (File) Kind() string {
	return "file"
}

func (f File) Equal(rhs interface{}) bool {
	return f == rhs
}

func (f File) Hash() uint32 {
	return hash.UIntPtr(f.inner.Fd())
}

func (f File) Repr(int) string {
	return fmt.Sprintf("<file{%s %p}>", parse.Quote(f.inner.Name()), f.inner)
}
