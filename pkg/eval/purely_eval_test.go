package eval_test

import (
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/testutil"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
)

func TestPurelyEvalCompound(t *testing.T) {
	home, cleanup := testutil.InTempHome()
	defer cleanup()

	var tests = []struct {
		code      string
		upto      int
		wantValue string
		wantBad   bool
	}{
		{code: "foobar", wantValue: "foobar"},
		{code: "'foobar'", wantValue: "foobar"},
		{code: "foo'bar'", wantValue: "foobar"},
		{code: "$x", wantValue: "bar"},
		{code: "foo$x", wantValue: "foobar"},
		{code: "foo$x", upto: 3, wantValue: "foo"},
		{code: "~", wantValue: home},
		{code: "~/foo", wantValue: home + "/foo"},
		{code: "$ns:x", wantValue: "foo"},

		{code: "$bad", wantBad: true},
		{code: "$ns:bad", wantBad: true},

		{code: "[abc]", wantBad: true},
		{code: "$y", wantBad: true},
		{code: "a[0]", wantBad: true},
		{code: "$@x", wantBad: true},
	}

	ev := NewEvaler()
	g := NsBuilder{
		"x": vars.NewReadOnly("bar"),
		"y": vars.NewReadOnly(vals.MakeList()),
	}.
		AddNs("ns", NsBuilder{"x": vars.NewReadOnly("foo")}.Ns()).
		Ns()
	ev.AddGlobal(g)

	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			n := &parse.Compound{}
			err := parse.ParseAs(
				parse.Source{Name: "[test]", Code: test.code}, n, nil)
			if err != nil {
				panic(err)
			}

			upto := test.upto
			if upto == 0 {
				upto = -1
			}
			value, ok := ev.PurelyEvalPartialCompound(n, upto)

			if value != test.wantValue {
				t.Errorf("got value %q, want %q", value, test.wantValue)
			}
			if ok != !test.wantBad {
				t.Errorf("got ok %v, want %v", ok, !test.wantBad)
			}
		})
	}
}
