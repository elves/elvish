package vals

import (
	"reflect"
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

// Kind tests the Kind of the value.
func (vt ValueTester) Kind(wantKind string) ValueTester {
	vt.t.Helper()
	kind := Kind(vt.v)
	if kind != wantKind {
		vt.t.Errorf("Kind(v) = %s, want %s", kind, wantKind)
	}
	return vt
}

// Bool tests the Boool of the value.
func (vt ValueTester) Bool(wantBool bool) ValueTester {
	vt.t.Helper()
	b := Bool(vt.v)
	if b != wantBool {
		vt.t.Errorf("Bool(v) = %v, want %v", b, wantBool)
	}
	return vt
}

// Hash tests the Hash of the value.
func (vt ValueTester) Hash(wantHash uint32) ValueTester {
	vt.t.Helper()
	hash := Hash(vt.v)
	if hash != wantHash {
		vt.t.Errorf("Hash(v) = %v, want %v", hash, wantHash)
	}
	return vt
}

// Len tests the Len of the value.
func (vt ValueTester) Len(wantLen int) ValueTester {
	vt.t.Helper()
	kind := Len(vt.v)
	if kind != wantLen {
		vt.t.Errorf("Len(v) = %v, want %v", kind, wantLen)
	}
	return vt
}

// Repr tests the Repr of the value.
func (vt ValueTester) Repr(wantRepr string) ValueTester {
	vt.t.Helper()
	kind := Repr(vt.v, NoPretty)
	if kind != wantRepr {
		vt.t.Errorf("Repr(v) = %s, want %s", kind, wantRepr)
	}
	return vt
}

// Equal tests that the value is Equal to every of the given values.
func (vt ValueTester) Equal(others ...interface{}) ValueTester {
	vt.t.Helper()
	for _, other := range others {
		eq := Equal(vt.v, other)
		if !eq {
			vt.t.Errorf("Equal(v, %v) = false, want true", other)
		}
	}
	return vt
}

// NotEqual tests that the value is not Equal to any of the given values.
func (vt ValueTester) NotEqual(others ...interface{}) ValueTester {
	vt.t.Helper()
	for _, other := range others {
		eq := Equal(vt.v, other)
		if eq {
			vt.t.Errorf("Equal(v, %v) = true, want false", other)
		}
	}
	return vt
}

// HasKey tests that the value has each of the given keys.
func (vt ValueTester) HasKey(keys ...interface{}) ValueTester {
	vt.t.Helper()
	for _, key := range keys {
		has := HasKey(vt.v, key)
		if !has {
			vt.t.Errorf("HasKey(v, %v) = false, want true", key)
		}
	}
	return vt
}

// HasNoKey tests that the value does not have any of the given keys.
func (vt ValueTester) HasNoKey(keys ...interface{}) ValueTester {
	vt.t.Helper()
	for _, key := range keys {
		has := HasKey(vt.v, key)
		if has {
			vt.t.Errorf("HasKey(v, %v) = true, want false", key)
		}
	}
	return vt
}

// AllKeys tests that the given keys match what the result of IterateKeys on the
// value.
//
// NOTE: This now checks equality using reflect.DeepEqual, since all the builtin
// types have string keys. This can be changed in future to use Equal is the
// need arises.
func (vt ValueTester) AllKeys(wantKeys ...interface{}) ValueTester {
	vt.t.Helper()
	keys, err := collectKeys(vt.v)
	if err != nil {
		vt.t.Errorf("IterateKeys(v, f) -> err %v, want nil", err)
	}
	if !reflect.DeepEqual(keys, wantKeys) {
		vt.t.Errorf("IterateKeys(v, f) calls f with %v, want %v", keys, wantKeys)
	}
	return vt
}

func collectKeys(v interface{}) ([]interface{}, error) {
	var keys []interface{}
	err := IterateKeys(v, func(k interface{}) bool {
		keys = append(keys, k)
		return true
	})
	return keys, err
}

// Index tests that Index'ing the value with the given key returns the wanted value
// and no error.
func (vt ValueTester) Index(key, wantVal interface{}) ValueTester {
	vt.t.Helper()
	got, err := Index(vt.v, key)
	if err != nil {
		vt.t.Errorf("Index(v, %v) -> err %v, want nil", key, err)
	}
	if !Equal(got, wantVal) {
		vt.t.Errorf("Index(v, %v) -> %v, want %v", key, got, wantVal)
	}
	return vt
}

// IndexError tests that Index'ing the value with the given key returns the given
// error.
func (vt ValueTester) IndexError(key interface{}, wantErr error) ValueTester {
	vt.t.Helper()
	_, err := Index(vt.v, key)
	if !reflect.DeepEqual(err, wantErr) {
		vt.t.Errorf("Index(v, %v) -> err %v, want %v", key, err, wantErr)
	}
	return vt
}

// Assoc tests that Assoc'ing the value with the given key-value pair returns
// the wanted new value and no error.
func (vt ValueTester) Assoc(key, val, wantNew interface{}) ValueTester {
	vt.t.Helper()
	got, err := Assoc(vt.v, key, val)
	if err != nil {
		vt.t.Errorf("Assoc(v, %v) -> err %v, want nil", key, err)
	}
	if !Equal(got, wantNew) {
		vt.t.Errorf("Assoc(v, %v) -> %v, want %v", key, got, wantNew)
	}
	return vt
}

// AssocError tests that Assoc'ing the value with the given key-value pair
// returns the given error.
func (vt ValueTester) AssocError(key, val interface{}, wantErr error) ValueTester {
	vt.t.Helper()
	_, err := Assoc(vt.v, key, val)
	if !reflect.DeepEqual(err, wantErr) {
		vt.t.Errorf("Assoc(v, %v) -> err %v, want %v", key, err, wantErr)
	}
	return vt
}

// Eq returns a tt.Matcher that matches using the Equal function.
func Eq(r interface{}) tt.Matcher { return equalMatcher{r} }

type equalMatcher struct{ want interface{} }

func (em equalMatcher) Match(got tt.RetValue) bool { return Equal(got, em.want) }
