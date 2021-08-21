package vals

import (
	"fmt"
	"os"

	"src.elv.sh/pkg/persistent/hash"
)

// Pipe wraps a pair of pointers to os.File that are the two ends of the same
// pipe.
type Pipe struct {
	ReadEnd, WriteEnd *os.File
}

var _ PseudoStructMap = Pipe{}

// NewPipe creates a new Pipe value.
func NewPipe(r, w *os.File) Pipe {
	return Pipe{r, w}
}

// Kind returns "pipe".
func (Pipe) Kind() string {
	return "pipe"
}

// Equal compares based on the equality of the two consistuent files.
func (p Pipe) Equal(rhs interface{}) bool {
	q, ok := rhs.(Pipe)
	if !ok {
		return false
	}
	return Equal(p.ReadEnd, q.ReadEnd) && Equal(p.WriteEnd, q.WriteEnd)
}

// Hash calculates the hash based on the two constituent files.
func (p Pipe) Hash() uint32 {
	return hash.DJB(Hash(p.ReadEnd), Hash(p.WriteEnd))
}

// Repr writes an opaque representation containing the FDs of the two
// constituent files.
func (p Pipe) Repr(int) string {
	return fmt.Sprintf("<pipe{%v %v}>", p.ReadEnd.Fd(), p.WriteEnd.Fd())
}

// Fields returns fields of the Pipe value.
func (p Pipe) Fields() StructMap {
	return pipeFields{p.ReadEnd, p.WriteEnd}
}

type pipeFields struct{ R, W *os.File }

func (pipeFields) IsStructMap() {}
