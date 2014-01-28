package parse

import (
	"fmt"
	"testing"
)

var atouRangeTests = []uintptr{
	0, 1, 2, ^uintptr(0),
}

var atouFormatTests = []struct {
	in     string
	wanted uintptr
}{
	{"012", 12},
}

var atouErrorTests = []string{
	"1e2", "foo", "1.9", "0x12",
}

func TestAtouRange(t *testing.T) {
	for _, wanted := range atouRangeTests {
		in := fmt.Sprint(wanted)
		out, err := Atou(in)
		if out != wanted || err != nil {
			t.Errorf("Atou(%q) => (%v, %v), want (%v, nil)", in, out, err, wanted)
		}
	}
}

func TestAtouFormat(t *testing.T) {
	for _, tt := range atouFormatTests {
		out, err := Atou(tt.in)
		if out != tt.wanted || err != nil {
			t.Errorf("Atou(%q) => (%v, %v), want (%v, nil)", tt.in, out, err, tt.wanted)
		}
	}
}

func TestAtouError(t *testing.T) {
	for _, in := range atouErrorTests {
		out, err := Atou(in)
		if err == nil {
			t.Errorf("Atou(%q) => (out = %v, err = %v), want err != nil", in, out, err)
		}
	}
}
