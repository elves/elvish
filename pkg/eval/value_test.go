package eval

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/glob"
)

var reprTests = []struct {
	v    interface{}
	want string
}{
	{"233", "233"},
	{"a\nb", `"a\nb"`},
	{"foo bar", "'foo bar'"},
	{"a\x00b", `"a\x00b"`},
	{true, "$true"},
	{false, "$false"},
	{&Exception{nil, nil}, "$ok"},
	{&Exception{errors.New("foo bar"), nil}, "?(fail 'foo bar')"},
	{&Exception{
		PipelineError{[]*Exception{{nil, nil}, {errors.New("lorem"), nil}}}, nil},
		"?(multi-error $ok ?(fail lorem))"},
	{&Exception{Return, nil}, "?(return)"},
	{vals.EmptyList, "[]"},
	{vals.MakeList("bash", false), "[bash $false]"},
	{vals.EmptyMap, "[&]"},
	{vals.MakeMap(&Exception{nil, nil}, "elvish"), "[&$ok=elvish]"},
	// TODO: test maps of more elements
}

func TestRepr(t *testing.T) {
	for _, test := range reprTests {
		repr := vals.Repr(test.v, vals.NoPretty)
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
