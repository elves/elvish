package vals

import (
	"fmt"
	"testing"

	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/persistent/hash"
)

func TestPipe(t *testing.T) {
	pr, pw := must.Pipe()
	defer pr.Close()
	defer pw.Close()

	TestValue(t, NewPipe(pr, pw)).
		Kind("pipe").
		Bool(true).
		Hash(hash.DJB(hash.UIntPtr(pr.Fd()), hash.UIntPtr(pw.Fd()))).
		Repr(fmt.Sprintf("<pipe{%v %v}>", pr.Fd(), pw.Fd())).
		Equal(NewPipe(pr, pw)).
		NotEqual(123, "a string", NewPipe(pw, pr)).
		AllKeys("r", "w").
		Index("r", pr).
		Index("w", pw)
}
