package sys

import (
	"syscall"
	"testing"
)

func mustNil(e error) {
	if e != nil {
		panic("error is not nil")
	}
}

func TestGetSetNonblock(t *testing.T) {
	var p [2]int
	mustNil(syscall.Pipe(p[:]))
	for _, b := range []bool{true, false} {
		if e := SetNonblock(p[0], b); e != nil {
			t.Errorf("SetNonblock(%v, %v) => %v, want <nil>", p[0], b, e)
		}
		if nb, e := GetNonblock(p[0]); nb != b || e != nil {
			t.Errorf("GetNonblock(%v) => (%v, %v), want (%v, <nil>)", p[0], nb, e, b)
		}
	}
	syscall.Close(p[0])
	syscall.Close(p[1])
	if e := SetNonblock(p[0], true); e == nil {
		t.Errorf("SetNonblock(%v, true) => <nil>, want non-<nil>", p[0])
	}
}
