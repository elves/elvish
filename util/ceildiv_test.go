package util

import "testing"

var ceilDivTests = []struct {
	a, b, out int
}{
	{9, 3, 3},
	{10, 3, 4},
	{11, 3, 4},
	{12, 3, 4},
}

func TestCeilDiv(t *testing.T) {
	for _, tt := range ceilDivTests {
		if o := CeilDiv(tt.a, tt.b); o != tt.out {
			t.Errorf("CeilDiv(%v, %v) => %v, want %v", tt.a, tt.b, o, tt.out)
		}
	}
}
