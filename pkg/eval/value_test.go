package eval

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/glob"
)

var reprTests = []struct {
	v    any
	want string
}{
	{"233", "233"},
	{"a\nb", `"a\nb"`},
	{"foo bar", "'foo bar'"},
	{"a\x00b", `"a\x00b"`},
	{true, "$true"},
	{false, "$false"},
	{vals.EmptyList, "[]"},
	{vals.MakeList("bash", false), "[bash $false]"},
	{vals.EmptyMap, "[&]"},
	{vals.MakeMap(&exception{nil, nil}, "elvish"), "[&$ok=elvish]"},
	// TODO: test maps of more elements
}

func TestRepr(t *testing.T) {
	for _, test := range reprTests {
		repr := vals.ReprPlain(test.v)
		if repr != test.want {
			t.Errorf("Repr = %s, want %s", repr, test.want)
		}
	}
}

var stringToSegmentsTests = []struct {
	s    string
	want []glob.Segment
}{
	{"", []glob.Segment{}},
	{"a", []glob.Segment{glob.Literal{Data: "a"}}},
	{"/a", []glob.Segment{glob.Slash{}, glob.Literal{Data: "a"}}},
	{"a/", []glob.Segment{glob.Literal{Data: "a"}, glob.Slash{}}},
	{"/a/", []glob.Segment{glob.Slash{}, glob.Literal{Data: "a"}, glob.Slash{}}},
	{"a//b", []glob.Segment{glob.Literal{Data: "a"}, glob.Slash{}, glob.Literal{Data: "b"}}},
}

func TestStringToSegments(t *testing.T) {
	for _, tc := range stringToSegmentsTests {
		segs := stringToSegments(tc.s)
		if !reflect.DeepEqual(segs, tc.want) {
			t.Errorf("stringToSegments(%q) => %v, want %v", tc.s, segs, tc.want)
		}
	}
}
