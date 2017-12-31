package eval

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/glob"
)

var reprTests = []struct {
	v    types.Value
	want string
}{
	{String("233"), "233"},
	{String("a\nb"), `"a\nb"`},
	{String("foo bar"), "'foo bar'"},
	{String("a\x00b"), `"a\x00b"`},
	{Bool(true), "$true"},
	{Bool(false), "$false"},
	{&Exception{nil, nil}, "$ok"},
	{&Exception{errors.New("foo bar"), nil}, "?(fail 'foo bar')"},
	{&Exception{
		PipelineError{[]*Exception{{nil, nil}, {errors.New("lorem"), nil}}}, nil},
		"?(multi-error $ok ?(fail lorem))"},
	{&Exception{Return, nil}, "?(return)"},
	{types.EmptyList, "[]"},
	{types.MakeList(String("bash"), Bool(false)), "[bash $false]"},
	{ConvertToMap(map[types.Value]types.Value{}), "[&]"},
	{ConvertToMap(map[types.Value]types.Value{&Exception{nil, nil}: String("elvish")}), "[&$ok=elvish]"},
	// TODO: test maps of more elements
}

func TestRepr(t *testing.T) {
	for _, test := range reprTests {
		repr := test.v.Repr(types.NoPretty)
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
	{"a", []glob.Segment{glob.Literal{"a"}}},
	{"/a", []glob.Segment{glob.Slash{}, glob.Literal{"a"}}},
	{"a/", []glob.Segment{glob.Literal{"a"}, glob.Slash{}}},
	{"/a/", []glob.Segment{glob.Slash{}, glob.Literal{"a"}, glob.Slash{}}},
	{"a//b", []glob.Segment{glob.Literal{"a"}, glob.Slash{}, glob.Literal{"b"}}},
}

func TestStringToSegments(t *testing.T) {
	for _, tc := range stringToSegmentsTests {
		segs := stringToSegments(tc.s)
		if !reflect.DeepEqual(segs, tc.want) {
			t.Errorf("stringToSegments(%q) => %v, want %v", tc.s, segs, tc.want)
		}
	}
}
