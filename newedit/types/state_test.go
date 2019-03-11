package types

import (
	"reflect"
	"testing"
)

func TestInsertAtDot(t *testing.T) {
	st := &State{Raw: RawState{Code: "ab", Dot: 1}}
	st.InsertAtDot("xy")
	wantRawState := RawState{Code: "axyb", Dot: 3}
	if !reflect.DeepEqual(st.Raw, wantRawState) {
		t.Errorf("got raw state %v, want %v", st.Raw, wantRawState)
	}
}
