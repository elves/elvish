package eval

import (
	"errors"
	"testing"
)

var reprTests = []struct {
	v    Value
	want string
}{
	{String("233"), "233"},
	{String("a\nb"), "'a\nb'"},
	{String("foo bar"), "'foo bar'"},
	{String("a\x00b"), `"a\x00b"`},
	{Bool(true), "$true"},
	{Bool(false), "$false"},
	{Error{nil}, "$ok"},
	{Error{errors.New("foo bar")}, "?(error 'foo bar')"},
	{Error{multiError{[]Error{{nil}, {errors.New("lorem")}}}},
		"?(multi-error $ok ?(error lorem))"},
	{Error{Return}, "?(return)"},
	{List{&[]Value{}}, "[]"},
	{List{&[]Value{String("bash"), Bool(false)}}, "[bash $false]"},
	{Map{&map[Value]Value{}}, "[&]"},
	{Map{&map[Value]Value{Error{nil}: String("elvish")}}, "[&$ok elvish]"},
	// TODO: test maps of more elements
}

func TestRepr(t *testing.T) {
	for _, test := range reprTests {
		repr := test.v.Repr()
		if repr != test.want {
			t.Errorf("Repr = %s, want %s", repr, test.want)
		}
	}
}
