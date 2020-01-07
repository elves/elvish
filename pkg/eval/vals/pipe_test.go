package vals

import (
	"fmt"
	"os"
	"testing"

	"github.com/xiaq/persistent/hash"
)

func TestPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	defer w.Close()

	TestValue(t, NewPipe(r, w)).
		HasKind("pipe").
		HasHash(hash.DJB(hash.UIntPtr(r.Fd()), hash.UIntPtr(w.Fd()))).
		HasRepr(fmt.Sprintf("<pipe{%v %v}>", r.Fd(), w.Fd())).
		IsEqualTo(NewPipe(r, w)).
		IsNotEqualTo(123, "a string", NewPipe(w, r))
}
