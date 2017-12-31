package eval

import (
	"testing"

	"github.com/elves/elvish/eval/types"
)

var (
	testStructDescriptor = NewStructDescriptor("foo", "bar")
	testStruct           = &Struct{testStructDescriptor, strs("lorem", "ipsum")}
	testStruct2          = &Struct{testStructDescriptor, strs("lorem", "dolor")}
)

func TestStructMethods(t *testing.T) {
	if l := testStruct.Len(); l != 2 {
		t.Errorf("testStruct.Len() = %d, want 2", l)
	}
	if foo := testStruct.IndexOne(String("foo")); foo != String("lorem") {
		t.Errorf(`testStruct.IndexOne("foo") = %q, want "lorem"`, foo)
	}
	if testStruct.Equal(testStruct2) {
		t.Errorf(`testStruct.Equal(testStruct2) => true, want false`)
	}
	if s2 := testStruct.Assoc(String("bar"), String("dolor")); !s2.Equal(testStruct2) {
		t.Errorf(`testStruct.Assoc(...) => %v, want %v`, s2, testStruct2)
	}
	wantRepr := "[&foo=lorem &bar=ipsum]"
	if gotRepr := testStruct.Repr(types.NoPretty); gotRepr != wantRepr {
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
