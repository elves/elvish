package vals

import (
	"fmt"
	"os"

	"github.com/xiaq/persistent/hash"
)

// Pipe wraps a pair of pointers to os.File that are the two ends of the same
// pipe.
type Pipe struct {
	ReadEnd, WriteEnd *os.File
}

var _ interface{} = Pipe{}

// NewPipe creates a new Pipe value.
func NewPipe(r, w *os.File) Pipe {
	return Pipe{r, w}
}

func (Pipe) Kind() string {
	return "pipe"
}

func (p Pipe) Equal(rhs interface{}) bool {
	return p == rhs
}

func (p Pipe) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.UIntPtr(p.ReadEnd.Fd()))
	h = hash.DJBCombine(h, hash.UIntPtr(p.WriteEnd.Fd()))
	return h
}

func (p Pipe) Repr(int) string {
	return fmt.Sprintf("<pipe{%v %v}>", p.ReadEnd.Fd(), p.WriteEnd.Fd())
}
