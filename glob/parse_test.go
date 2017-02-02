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
	{`foo`, []Segment{Literal{"foo"}}},
	{`*foo*bar`, []Segment{
		Wild{Star, false, nil}, Literal{"foo"},
		Wild{Star, false, nil}, Literal{"bar"}}},
	{`foo**bar`, []Segment{
		Literal{"foo"}, Wild{StarStar, false, nil}, Literal{"bar"}}},
	{`/usr/a**b/c`, []Segment{
		Slash{}, Literal{"usr"}, Slash{}, Literal{"a"},
		Wild{StarStar, false, nil}, Literal{"b"}, Slash{}, Literal{"c"}}},
	{`??b`, []Segment{
		Wild{Question, false, nil}, Wild{Question, false, nil}, Literal{"b"}}},
	// Multiple slashes should be parsed as one.
	{`//a//b`, []Segment{
		Slash{}, Literal{"a"}, Slash{}, Literal{"b"}}},
	// Escaping.
	{`\*\?b`, []Segment{
		Literal{"*?b"},
	}},
	{`abc\`, []Segment{
		Literal{"abc"},
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
