package vals

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
	"github.com/xiaq/persistent/hash"
)

// File wraps a pointer to os.File.
type File struct {
	Inner *os.File
}

var _ interface{} = File{}

// NewFile creates a new File value.
func NewFile(inner *os.File) File {
	return File{inner}
}

func (File) Kind() string {
	return "file"
}

func (f File) Equal(rhs interface{}) bool {
	return f == rhs
}

func (f File) Hash() uint32 {
	return hash.UIntPtr(f.Inner.Fd())
}

func (f File) Repr(int) string {
	return fmt.Sprintf("<file{%s %p}>", parse.Quote(f.Inner.Name()), f.Inner)
}
