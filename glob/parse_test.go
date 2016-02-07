package glob

import (
	"reflect"
	"testing"
)

var parseCases = []struct {
	src  string
	want Pattern
}{
	{``, Pattern{[]Segment{}}},
	{`foo`, Pattern{[]Segment{{Literal, "foo"}}}},
	{`*foo*bar`, Pattern{[]Segment{
		{Star, ""}, {Literal, "foo"}, {Star, ""}, {Literal, "bar"}}}},
	{`foo**bar`, Pattern{[]Segment{
		{Literal, "foo"}, {StarStar, ""}, {Literal, "bar"}}}},
	{`/usr/a**b/c`, Pattern{[]Segment{
		{Slash, ""}, {Literal, "usr"}, {Slash, ""}, {Literal, "a"},
		{StarStar, ""}, {Literal, "b"}, {Slash, ""}, {Literal, "c"}}}},
}

func TestParse(t *testing.T) {
	for _, tc := range parseCases {
		p := Parse(tc.src)
		if !reflect.DeepEqual(p, tc.want) {
			t.Errorf("Parse(%q) => %v, want %v", p, tc.want)
		}
	}
}
