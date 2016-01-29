package print

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

var deepTests = []struct {
	in     interface{}
	wanted string
}{
	{1, "1"},
	{"foobar", `"foobar"`},
	{map[string]int{"foobar": 42}, `map[string]int{"foobar": 42}`},
	{[]int{42, 44}, `[]int{42, 44}`},
	{[]int(nil), `nil`},
	{(*int)(nil), `nil`},
	{&S{42, "DON'T PANIC", &T{map[string]string{"foo": "bar"}}, G{}},
		`&print.S{I: 42, S: "DON'T PANIC", Pt: &print.T{M: map[string]string{"foo": "bar"}}, G: <G>}`},
	{[]interface{}{&U{42, "DON'T PANIC"}, 42, "DON'T PANIC"}, `[]interface {}{&print.U{I: 42, S: "DON'T PANIC"}, 42, "DON'T PANIC"}`},
}

func TestDeep(t *testing.T) {
	for _, tt := range deepTests {
		if out := Deeply(tt.in); out != tt.wanted {
			t.Errorf("Deep(%v) => %#q, want %#q", tt.in, out, tt.wanted)
		}
	}
}
