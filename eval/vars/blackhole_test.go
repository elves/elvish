package vars

import (
	"testing"

	"github.com/elves/elvish/tt"
)

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
	tt.Test(t, tt.Fn("IsBlackhole", IsBlackhole), tt.Table{
		tt.Args(NewBlackhole()).Rets(true),
		tt.Args(FromInit("")).Rets(false),
	})
}
