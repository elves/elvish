package vars

import (
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/tt"
)

func TestNewReadOnly(t *testing.T) {
	v := NewReadOnly("haha")
	if v.Get() != "haha" {
		t.Errorf("Get doesn't return initial value")
	}

	err := v.Set("lala")
	if _, ok := err.(errs.SetReadOnlyVar); !ok {
		t.Errorf("Set a readonly var doesn't error as expected: %#v", err)
	}
}

func TestIsReadOnly(t *testing.T) {
	tt.Test(t, IsReadOnly,
		Args(NewReadOnly("foo")).Rets(true),
		Args(FromGet(func() any { return "foo" })).Rets(true),
		Args(FromInit("foo")).Rets(false),
	)
}
