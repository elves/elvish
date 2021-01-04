package eval_test

import (
	"reflect"
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
		wantErr   error
	}{
		{code: "foobar", wantValue: "foobar"},
		{code: "'foobar'", wantValue: "foobar"},
		{code: "foo'bar'", wantValue: "foobar"},
		{code: "$x", wantValue: "bar"},
		{code: "foo$x", wantValue: "foobar"},
		{code: "foo$x", upto: 3, wantValue: "foo"},
		{code: "~", wantValue: home},
		{code: "~/foo", wantValue: home + "/foo"},

		{code: "[abc]", wantErr: ErrImpure},
		{code: "$y", wantErr: ErrImpure},
		{code: "a[0]", wantErr: ErrImpure},
		{code: "$@x", wantErr: ErrImpure},
	}

	ev := NewEvaler()
	ev.SetGlobal(NsBuilder{
		"x": vars.NewReadOnly("bar"),
		"y": vars.NewReadOnly(vals.MakeList()),
	}.Ns())

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
			value, err := ev.PurelyEvalPartialCompound(n, upto)

			if value != test.wantValue {
				t.Errorf("got value %q, want %q", value, test.wantValue)
			}
			if !reflect.DeepEqual(err, test.wantErr) {
				t.Errorf("got error %v, want %q", err, test.wantErr)
			}
		})
	}
}
