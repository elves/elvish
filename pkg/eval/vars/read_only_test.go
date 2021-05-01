package vars

import (
	"testing"

	"src.elv.sh/pkg/eval/errs"
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
