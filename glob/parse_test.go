package glob

import (
	"reflect"
	"testing"
)

var parseCases = []struct {
	src  string
	want []Segment
}{
	{``, []Segment{}},
	{`foo`, []Segment{{Literal, "foo"}}},
	{`*foo*bar`, []Segment{
		{Star, ""}, {Literal, "foo"}, {Star, ""}, {Literal, "bar"}}},
	{`foo**bar`, []Segment{
		{Literal, "foo"}, {StarStar, ""}, {Literal, "bar"}}},
	{`/usr/a**b/c`, []Segment{
		{Slash, ""}, {Literal, "usr"}, {Slash, ""}, {Literal, "a"},
		{StarStar, ""}, {Literal, "b"}, {Slash, ""}, {Literal, "c"}}},
	{`??b`, []Segment{
		{Question, ""}, {Question, ""}, {Literal, "b"}}},
	// Multiple slashes should be parsed as one.
	{`//a//b`, []Segment{
		{Slash, ""}, {Literal, "a"}, {Slash, ""}, {Literal, "b"}}},
	// Escaping.
	{`\*\?b`, []Segment{
		{Literal, "*?b"},
	}},
	{`abc\`, []Segment{
		{Literal, "abc"},
	}},
}

func TestParse(t *testing.T) {
	for _, tc := range parseCases {
		p := Parse(tc.src)
		if !reflect.DeepEqual(p.Segments, tc.want) {
			t.Errorf("Parse(%q) => %v, want %v", tc.src, p, tc.want)
		}
	}
}
