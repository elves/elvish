package eval

import (
	"fmt"
	"os"

	"github.com/xiaq/persistent/hash"
)

type Pipe struct {
	r, w *os.File
}

var _ Value = Pipe{}

func (Pipe) Kind() string {
	return "pipe"
}

func (p Pipe) Equal(rhs interface{}) bool {
	return p == rhs
}

func (p Pipe) Hash() uint32 {
	h := hash.DJBInit
	h = hash.DJBCombine(h, hash.UIntPtr(p.r.Fd()))
	h = hash.DJBCombine(h, hash.UIntPtr(p.w.Fd()))
	return h
}

func (p Pipe) Repr(int) string {
	return fmt.Sprintf("<pipe{%v %v}>", p.r.Fd(), p.w.Fd())
}
