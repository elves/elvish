package eval

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/tt"
	"github.com/xiaq/persistent/hash"
)

func TestCause(t *testing.T) {
	err := errors.New("ordinary error")
	tt.Test(t, tt.Fn("Cause", Cause), tt.Table{
		tt.Args(err).Rets(err),
		tt.Args(makeException(err)).Rets(err),
	})
}

func TestException(t *testing.T) {
	err := errors.New("error")
	exc := makeException(err)
	vals.TestValue(t, exc).
		Kind("exception").
		Bool(false).
		Hash(hash.Pointer(unsafe.Pointer(exc))).
		Equal(exc).
		NotEqual(makeException(errors.New("error"))).
		AllKeys("cause").
		Index("cause", err).
		IndexError("stack", vals.NoSuchKey("stack")).
		Repr("?(fail error)")

	vals.TestValue(t, OK).
		Kind("exception").
		Bool(true).
		Repr("$ok")
}

func makeException(cause error, entries ...*diag.Context) *Exception {
	var s *stackTrace
	for i := len(entries) - 1; i >= 0; i-- {
		s = &stackTrace{head: entries[i], next: s}
	}
	return &Exception{cause, s}
}

func TestErrors(t *testing.T) {
	tt.Test(t, tt.Fn("Error", error.Error), tt.Table{
		tt.Args(Return).Rets("return"),
		tt.Args(Break).Rets("break"),
		tt.Args(Continue).Rets("continue"),
	})
}
