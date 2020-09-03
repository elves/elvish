package eval

import (
	"errors"
	"runtime"
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
	err := FailError{"error"}
	exc := makeException(err)
	vals.TestValue(t, exc).
		Kind("exception").
		Bool(false).
		Hash(hash.Pointer(unsafe.Pointer(exc))).
		Equal(exc).
		NotEqual(makeException(errors.New("error"))).
		AllKeys("reason").
		Index("reason", err).
		IndexError("stack", vals.NoSuchKey("stack")).
		Repr("[&reason=[&content=error &type=fail]]")

	vals.TestValue(t, OK).
		Kind("exception").
		Bool(true).
		Repr("$ok")
}

func makeException(cause error, entries ...*diag.Context) *Exception {
	var s *StackTrace
	for i := len(entries) - 1; i >= 0; i-- {
		s = &StackTrace{Head: entries[i], Next: s}
	}
	return &Exception{cause, s}
}

func TestFlow_Fields(t *testing.T) {
	Test(t,
		That("put ?(return)[reason][type name]").Puts("flow", "return"),
	)
}

func TestExternalCmdExit_Fields(t *testing.T) {
	badCmd := "false"
	if runtime.GOOS == "windows" {
		badCmd = "cmd /c exit 1"
	}
	Test(t,
		That("put ?("+badCmd+")[reason][type exit-status]").
			Puts("external-cmd/exited", "1"),
		// TODO: Test killed and stopped commands
	)
}

func TestPipelineError_Fields(t *testing.T) {
	Test(t,
		That("put ?(fail 1 | fail 2)[reason][type]").Puts("pipeline"),
		That("count ?(fail 1 | fail 2)[reason][exceptions]").Puts("2"),
		That("put ?(fail 1 | fail 2)[reason][exceptions][0][reason][type]").
			Puts("fail"),
	)
}

func TestErrorMethods(t *testing.T) {
	tt.Test(t, tt.Fn("Error", error.Error), tt.Table{
		tt.Args(makeException(errors.New("err"))).Rets("err"),

		tt.Args(makePipelineError([]*Exception{
			makeException(errors.New("err1")),
			makeException(errors.New("err2"))})).Rets("(err1 | err2)"),

		tt.Args(Return).Rets("return"),
		tt.Args(Break).Rets("break"),
		tt.Args(Continue).Rets("continue"),
		tt.Args(Flow(1000)).Rets("!(BAD FLOW: 1000)"),
	})
}
