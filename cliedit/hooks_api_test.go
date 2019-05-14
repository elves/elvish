package cliedit

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
)

func TestInitBeforeReadline(t *testing.T) {
	variable, cb := initBeforeReadline(eval.NewEvaler())
	called := 0
	variable.Set(vals.EmptyList.Cons(eval.NewGoFn("[test]", func() {
		called++
	})))
	cb()
	if called != 1 {
		t.Errorf("Called %d times, want once", called)
	}
	// TODO: Test input and output
}

func TestInitAfterReadline(t *testing.T) {
	variable, cb := initAfterReadline(eval.NewEvaler())
	called := 0
	calledWith := ""
	variable.Set(vals.EmptyList.Cons(eval.NewGoFn("[test]", func(s string) {
		called++
		calledWith = s
	})))
	cb("code")
	if called != 1 {
		t.Errorf("Called %d times, want once", called)
	}
	if calledWith != "code" {
		t.Errorf("Called with %q, want %q", calledWith, "code")
	}
	// TODO: Test input and output
}
