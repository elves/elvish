package types

import (
	"testing"
)

var (
	testStructDescriptor = NewStructDescriptor("foo", "bar")
	testStruct           = NewStruct(testStructDescriptor, []Value{String("lorem"), String("ipsum")})
	testStruct2          = NewStruct(testStructDescriptor, []Value{String("lorem"), String("dolor")})
)

func TestStructMethods(t *testing.T) {
	if l := testStruct.Len(); l != 2 {
		t.Errorf("testStruct.Len() = %d, want 2", l)
	}
	if foo, err := testStruct.Index(String("foo")); foo != String("lorem") {
		t.Errorf(`testStruct.Index("foo") = %q, want "lorem"`, foo)
	} else if err != nil {
		t.Errorf(`testStruct.Index("foo") => error %s, want no error`, err)
	}
	if testStruct.Equal(testStruct2) {
		t.Errorf(`testStruct.Equal(testStruct2) => true, want false`)
	}
	if s2, err := testStruct.Assoc(String("bar"), String("dolor")); !s2.Equal(testStruct2) {
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
