package completion

import (
	"reflect"
	"strings"
	"testing"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
)

func TestSortRawCandidates(t *testing.T) {
	ev := eval.NewEvaler()
	defer ev.Close()

	fakeDirSort := eval.NewBuiltinFn("test:fakeDirSort",
		func(fm *eval.Frame, opts eval.RawOptions, seed, a, b string) {
			out := fm.OutputChan()
			da, db := strings.HasSuffix(a, "/"), strings.HasSuffix(b, "/")
			isBefore := false

			switch {
			case da && !db:
				isBefore = true
			case !da && db:
				isBefore = false
			default:
				pa := strings.HasPrefix(a, seed)
				if pa {
					isBefore = true
				} else {
					isBefore = false
				}
			}
			out <- vals.Bool(isBefore)
		})

	for name, test := range map[string]struct {
		src    []string
		sorter eval.Callable
		seed   string
		want   []string
	}{
		"default": {
			src:  []string{"z", "y", "x"},
			want: []string{"x", "y", "z"},
		},
		"fakeDirSort": {
			sorter: fakeDirSort,
			seed:   "Y",
			src:    []string{"x", "z/", "y", "Y/"},
			want:   []string{"Y/", "z/", "x", "y"},
		},
	} {
		inputs := make([]rawCandidate, len(test.src))
		for i, s := range test.src {
			inputs[i] = plainCandidate(s)
		}

		result, err := sortRawCandidates(ev, test.sorter, test.seed, inputs)
		if err != nil {
			t.Errorf("test %s: got unexpected error %v", name, err)
		}

		got := make([]string, len(result))
		for i, r := range result {
			got[i] = r.text()
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %s: got unexpected result: %v, want %v", name, got, test.want)
		}
	}
}
