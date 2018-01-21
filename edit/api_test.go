package edit

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
)

func TestBuiltinFn(t *testing.T) {
	called := false
	builtinFn := &BuiltinFn{"foobar", func(*Editor) {
		if called {
			t.Errorf("builtin impl called multiple times, called not reset")
		}
		called = true
	}}

	if kind := builtinFn.Kind(); kind != "fn" {
		t.Errorf("Kind of BuiltinFn should be fn, is %q", kind)
	}
	if repr := builtinFn.Repr(10); repr != "$foobar" {
		t.Errorf("Repr of BuiltinFn should be $foobar, is %q", repr)
	}

	ec := &eval.Frame{Evaler: &eval.Evaler{}}

	if builtinFn.Call(ec, nil, nil) != errEditorInvalid {
		t.Errorf("BuiltinFn should error when Editor is nil, didn't")
	}

	ec.Editor = &Editor{active: false}
	if builtinFn.Call(ec, nil, nil) != errEditorInactive {
		t.Errorf("BuiltinFn should error when Editor is inactive, didn't")
	}

	ec.Editor = &Editor{active: true}

	if !util.Throws(func() {
		builtinFn.Call(ec, []types.Value{types.String("2")}, nil)
	}, eval.ErrNoArgAccepted) {
		t.Errorf("BuiltinFn should error when argument was supplied, didn't")
	}

	if !util.Throws(func() {
		builtinFn.Call(ec, nil, map[string]types.Value{"a": types.String("b")})
	}, eval.ErrNoOptAccepted) {
		t.Errorf("BuiltinFn should error when option was supplied, didn't")
	}

	builtinFn.Call(ec, nil, nil)
	if !called {
		t.Errorf("BuiltinFn should call its implementation, didn't")
	}
}
