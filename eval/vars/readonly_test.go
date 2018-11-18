package vars

import "testing"

func TestNewReadOnly(t *testing.T) {
	v := NewReadOnly("haha")
	if v.Get() != "haha" {
		t.Errorf("Get doesn't return initial value")
	}
	if v.Set("lala") != errSetReadOnlyVar {
		t.Errorf("Set doesn't error")
	}
}
