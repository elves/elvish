package util

import (
	"testing"
)

type S struct {
	I  int
	S  string
	Pt *T
	G  G
}

type T struct {
	M map[string]string
}

type G struct {
}

type U struct {
	I int
	S string
}

func (g G) GoString() string {
	return "<G>"
}

var deepPrintTests = []struct {
	in     interface{}
	wanted string
}{
	{1, "1"},
	{"foobar", `"foobar"`},
	{[]int{42, 44}, `[]int{42, 44}`},
	{[]int(nil), `nil`},
	{(*int)(nil), `nil`},
	{&S{42, "DON'T PANIC", &T{map[string]string{"foo": "bar"}}, G{}},
		`&util.S{I: 42, S: "DON'T PANIC", Pt: &util.T{M: map[string]string{"foo": "bar"}}, G: <G>}`},
	{[]interface{}{&U{42, "DON'T PANIC"}, 42, "DON'T PANIC"}, `[]interface {}{&util.U{I: 42, S: "DON'T PANIC"}, 42, "DON'T PANIC"}`},
}

func TestDeepPrint(t *testing.T) {
	for _, tt := range deepPrintTests {
		if out := DeepPrint(tt.in); out != tt.wanted {
			t.Errorf("DeepPrint(%v) => %#q, want %#q", tt.in, out, tt.wanted)
		}
	}
	// Test map.
	in := map[string]int{"foo": 42, "bar": 233}
	out := DeepPrint(in)
	wanted1 := `map[string]int{"foo": 42, "bar": 233}`
	wanted2 := `map[string]int{"bar": 233, "foo": 42}`
	if out != wanted1 && out != wanted2 {
		t.Errorf("DeepPrint(%v) => %#q, want %#q or %#q", in, out, wanted1, wanted2)
	}
}
