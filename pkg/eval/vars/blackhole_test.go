package vars

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestBlackhole(t *testing.T) {
	v := NewBlackhole()
	err := v.Set("foo")
	if err != nil {
		t.Errorf("v.Set(%q) -> %v, want nil", "foo", err)
	}
	val := v.Get()
	if val != nil {
		t.Errorf("v.Get() -> %v, want nil", val)
	}
}

func TestIsBlackhole(t *testing.T) {
	tt.Test(t, IsBlackhole,
		Args(NewBlackhole()).Rets(true),
		Args(FromInit("")).Rets(false),
	)
}
