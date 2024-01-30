package eval_test

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"src.elv.sh/pkg/diag"
	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/testutil"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/tt"
)

func TestReason(t *testing.T) {
	err := errors.New("ordinary error")
	tt.Test(t, Reason,
		Args(err).Rets(err),
		Args(makeException(err)).Rets(err),
	)
}

func TestException(t *testing.T) {
	err := FailError{"error"}
	exc := makeException(err)
	vals.TestValue(t, exc).
		Kind("exception").
		Bool(false).
		Hash(hash.Pointer(unsafe.Pointer(reflect.ValueOf(exc).Pointer()))).
		Equal(exc).
		NotEqual(makeException(errors.New("error"))).
		AllKeys("reason", "stack-trace").
		Index("reason", err).
		IndexError("stack", vals.NoSuchKey("stack")).
		Repr("[^exception &reason=[^fail-error &content=error &type=fail] &stack-trace=<...>]")

	vals.TestValue(t, OK).
		Kind("exception").
		Bool(true).
		Repr("$ok")
}

func TestException_Show(t *testing.T) {
	for _, p := range []*string{
		ExceptionCauseStartMarker, ExceptionCauseEndMarker,
		&diag.ContextBodyStartMarker, &diag.ContextBodyEndMarker} {

		testutil.Set(t, p, "")
	}

	tt.Test(t, Exception.Show,
		It("supports exceptions with one traceback frame").
			Args(makeException(
				errors.New("internal error"),
				diag.NewContext("a.elv", "echo bad", diag.Ranging{From: 5, To: 8})), "").
			Rets(Dedent(`
				Exception: internal error
				  a.elv:1:6-8: echo bad`)),

		It("supports exceptions with multiple traceback frames").
			Args(makeException(
				errors.New("internal error"),
				diag.NewContext("a.elv", "echo bad", diag.Ranging{From: 5, To: 8}),
				diag.NewContext("b.elv", "use foo", diag.Ranging{From: 0, To: 7})), "").
			Rets(Dedent(`
				Exception: internal error
				  a.elv:1:6-8: echo bad
				  b.elv:1:1-7: use foo`)),

		It("supports traceback frames with multi-line body text").
			Args(makeException(
				errors.New("internal error"),
				diag.NewContext("a.elv", "echo ba\nd", diag.Ranging{From: 5, To: 9})), "").
			Rets(Dedent(`
				Exception: internal error
				  a.elv:1:6-2:1:
				    echo ba
				    d`)),
	)
}

func makeException(cause error, entries ...*diag.Context) Exception {
	return NewException(cause, makeStackTrace(entries...))
}

// Creates a new StackTrace, using the first entry as the head.
func makeStackTrace(entries ...*diag.Context) *StackTrace {
	var s *StackTrace
	for i := len(entries) - 1; i >= 0; i-- {
		s = &StackTrace{Head: entries[i], Next: s}
	}
	return s
}

func TestErrorMethods(t *testing.T) {
	tt.Test(t, error.Error,
		Args(makeException(errors.New("err"))).Rets("err"),

		Args(MakePipelineError([]Exception{
			makeException(errors.New("err1")),
			makeException(errors.New("err2"))})).Rets("(err1 | err2)"),

		Args(Return).Rets("return"),
		Args(Break).Rets("break"),
		Args(Continue).Rets("continue"),
		Args(Flow(1000)).Rets("!(BAD FLOW: 1000)"),
	)
}
