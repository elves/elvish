package completion

import (
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
)

func TestRawFilterCandidates(t *testing.T) {
	passAll := eval.NewBuiltinFn("test:passAll",
		func(fm *eval.Frame, opts eval.RawOptions, pattern string, inputs eval.Inputs) {
			out := fm.OutputChan()
			inputs(func(v interface{}) {
				out <- vals.Bool(true)
			})
		})
	blockAll := eval.NewBuiltinFn("test:blockAll",
		func(fm *eval.Frame, opts eval.RawOptions, pattern string, inputs eval.Inputs) {
			out := fm.OutputChan()
			inputs(func(v interface{}) {
				out <- vals.Bool(false)
			})
		})

	tests := []filterRawCandidatesTest{
		{
			name:    "passAll",
			matcher: passAll,
			src:     []string{"1", "2", "3"},
			want:    []string{"1", "2", "3"},
		},
		{
			name:    "blockAll",
			matcher: blockAll,
			src:     []string{"1", "2", "3"},
			want:    []string{},
		},
	}

	testRawFilterCandidates(t, tests)
}
