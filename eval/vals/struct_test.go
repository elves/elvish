package vals

import (
	"testing"
)

var (
	testStructDescriptor = NewStructDescriptor("foo", "bar")
	testStruct           = NewStruct(testStructDescriptor, []interface{}{"lorem", "ipsum"})
	testStruct2          = NewStruct(testStructDescriptor, []interface{}{"lorem", "dolor"})
)

func TestStructMethods(t *testing.T) {
	if l := testStruct.Len(); l != 2 {
		t.Errorf("testStruct.Len() = %d, want 2", l)
	}
	if foo, ok := testStruct.Get("foo"); foo != "lorem" {
		t.Errorf(`testStruct.Get("foo") = %q, want "lorem"`, foo)
	} else if !ok {
		t.Errorf(`testStruct.Get("foo") => false, want true`)
	}
	if testStruct.Equal(testStruct2) {
		t.Errorf(`testStruct.Equal(testStruct2) => true, want false`)
	}
	if s2, err := testStruct.Assoc("bar", "dolor"); !Equal(s2, testStruct2) {
		t.Errorf(`testStruct.Assoc(...) => %v, want %v`, s2, testStruct2)
	} else if err != nil {
		t.Errorf(`testStruct.Assoc(...) => error %s, want no error`, err)
	}
	wantRepr := "[&foo=lorem &bar=ipsum]"
	if gotRepr := testStruct.Repr(NoPretty); gotRepr != wantRepr {
		t.Errorf(`testStruct.Repr() => %q, want %q`, gotRepr, wantRepr)
	}
	wantJSON := `{"foo":"lorem","bar":"ipsum"}`
	gotJSONBytes, err := testStruct.MarshalJSON()
	gotJSON := string(gotJSONBytes)
	if err != nil {
		t.Errorf(`testStruct.MarshalJSON() => error %v`, err)
	}
	if wantJSON != gotJSON {
		t.Errorf(`testStruct.MarshalJSON() => %q, want %q`, gotJSON, wantJSON)
	}
}
