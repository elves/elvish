package types

import (
	"testing"
)

var convertListIndexTests = []struct {
	name string
	// input
	expr string
	len  int
	// output
	wantOut *ListIndex
	wantErr bool
}{
	{name: "stringIndex", expr: "a", len: 0,
		wantErr: true},
	{name: "floatIndex", expr: "1.0", len: 0,
		wantErr: true},
	{name: "emptyZeroIndex", expr: "0", len: 0,
		wantErr: true},
	{name: "emptyPosIndex", expr: "1", len: 0,
		wantErr: true},
	{name: "emptyNegIndex", expr: "-1", len: 0,
		wantErr: true},
	// BUG(xiaq): Should not be error
	{name: "emptySliceAbbrevBoth", expr: ":", len: 0,
		wantErr: true},
	{name: "i<-n", expr: "-2", len: 1, wantErr: true},
	{name: "i=-n", expr: "-1", len: 1,
		wantOut: &ListIndex{Lower: 0}},
	{name: "-n<i<0", expr: "-1", len: 2,
		wantOut: &ListIndex{Lower: 1}},
	{name: "i=0", expr: "0", len: 2,
		wantOut: &ListIndex{Lower: 0}},
	{name: "0<i<n", expr: "1", len: 2,
		wantOut: &ListIndex{Lower: 1}},
	{name: "i=n", expr: "1", len: 1,
		wantErr: true},
	{name: "i>n", expr: "2", len: 1,
		wantErr: true},
	{name: "sliceAbbrevBoth", expr: ":", len: 1,
		wantOut: &ListIndex{Slice: true, Lower: 0, Upper: 1}},
	{name: "sliceAbbrevBegin", expr: ":1", len: 1,
		wantOut: &ListIndex{Slice: true, Lower: 0, Upper: 1}},
	{name: "sliceAbbrevEnd", expr: "0:", len: 1,
		wantOut: &ListIndex{Slice: true, Lower: 0, Upper: 1}},
	{name: "sliceNegEnd", expr: "0:-1", len: 1,
		wantOut: &ListIndex{Slice: true, Lower: 0, Upper: 0}},
	{name: "sliceBeginEqualEnd", expr: "1:1", len: 2,
		wantOut: &ListIndex{Slice: true, Lower: 1, Upper: 1}},
	{name: "sliceBeginAboveEnd", expr: "1:0", len: 2,
		wantErr: true},
}

func TestConvertListIndex(t *testing.T) {
	for _, test := range convertListIndexTests {
		index, err := ConvertListIndex(test.expr, test.len)
		if !eqListIndex(index, test.wantOut) {
			t.Errorf("ConvertListIndex(%q, %d) => %v, want %v",
				test.expr, test.len, index, test.wantOut)
		}
		wantErr := "no error"
		if test.wantErr {
			wantErr = "non-nil error"
		}
		if test.wantErr != (err != nil) {
			t.Errorf("ConvertListIndex(%q, %d) => err %v, want %s",
				err, wantErr)
		}
	}
}

func eqListIndex(a, b *ListIndex) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
