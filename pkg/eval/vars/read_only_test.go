package vars

import (
	"testing"
)

func TestNewReadOnly(t *testing.T) {
	v := NewReadOnly("v", "haha")
	if v.Get() != "haha" {
		t.Errorf("Get doesn't return initial value")
	}

	err := v.Set("lala")
	switch err := err.(type) {
	case *ErrSetReadOnlyVar:
		if err.VarName != "v" {
			t.Errorf("Set doesn't correctly report read-only error: expected err.VarName %v got %v",
				err.VarName, "v")
		}
	default:
		t.Errorf("Set doesn't report read-only error")
	}
}
