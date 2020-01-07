package vals

import (
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

// ValueTester is a helper for testing properties of a value.
type ValueTester struct {
	t *testing.T
	v interface{}
}

// TestValue returns a ValueTester.
func TestValue(t *testing.T, v interface{}) ValueTester {
	return ValueTester{t, v}
}

// HasKind tests the Kind of the value.
func (vt ValueTester) HasKind(wantKind string) ValueTester {
	vt.t.Helper()
	kind := Kind(vt.v)
	if kind != wantKind {
		vt.t.Errorf("Kind(v) = %s, want %s", kind, wantKind)
	}
	return vt
}

// HasHash tests the Hash of the value.
func (vt ValueTester) HasHash(wantHash uint32) ValueTester {
	vt.t.Helper()
	kind := Hash(vt.v)
	if kind != wantHash {
		vt.t.Errorf("Hash(v) = %v, want %v", kind, wantHash)
	}
	return vt
}

// HasRepr tests the Repr of the value.
func (vt ValueTester) HasRepr(wantRepr string) ValueTester {
	vt.t.Helper()
	kind := Repr(vt.v, -1)
	if kind != wantRepr {
		vt.t.Errorf("Repr(v) = %s, want %s", kind, wantRepr)
	}
	return vt
}

// IsEqualTo tests that the value is Equal to another.
func (vt ValueTester) IsEqualTo(others ...interface{}) ValueTester {
	vt.t.Helper()
	for _, other := range others {
		eq := Equal(vt.v, other)
		if !eq {
			vt.t.Errorf("Equal(v, %v) = false, want true", other)
		}
	}
	return vt
}

// IsNotEqualTo tests that the value is not Equal to another.
func (vt ValueTester) IsNotEqualTo(others ...interface{}) ValueTester {
	vt.t.Helper()
	for _, other := range others {
		eq := Equal(vt.v, other)
		if eq {
			vt.t.Errorf("Equal(v, %v) = true, want false", other)
		}
	}
	return vt
}

// Eq returns a tt.Matcher that matches using the Equal function.
func Eq(r interface{}) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want interface{} }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }
