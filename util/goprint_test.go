package util

import (
	"testing"
)

type S struct {
	I  int
	S  string
	Pt *T
	G G
}

type T struct {
	M map[string]string
}

type G struct {
}

func (g G) GoString() string {
	return "<G>"
}

var goPrintTests = []struct {
	in     interface{}
	wanted string
}{
	{1, "1"},
	{"foobar", `"foobar"`},
	{map[string]int{"foobar": 42}, `map[string]int{"foobar": 42}`},
	{[]int{42, 44}, `[]int{42, 44}`},
	{[]int(nil), `[]int(nil)`},
	{(*int)(nil), `*int(nil)`},
	{&S{42, "DON'T PANIC", &T{map[string]string{"foo": "bar"}}, G{}},
		`&util.S{I: 42, S: "DON'T PANIC", Pt: &util.T{M: map[string]string{"foo": "bar"}}, G: <G>}`},
}

func TestGoPrint(t *testing.T) {
	for _, tt := range goPrintTests {
		if out := GoPrint(tt.in); out != tt.wanted {
			t.Errorf("GoPrint(%v) => %#q, want %#q", tt.in, out, tt.wanted)
		}
	}
}
