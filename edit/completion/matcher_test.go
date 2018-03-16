package completion

import (
	"testing"
)

func TestBuiltinMatcher(t *testing.T) {
	tests := []filterRawCandidatesTest{
		{
			name:    "matchPrefix",
			matcher: matchPrefix,
			seed:    "x",
			src:     []string{"x1", "x2", "3"},
			want:    []string{"x1", "x2"},
		},
		{
			name:    "matchSubstr",
			matcher: matchSubstr,
			seed:    "x",
			src:     []string{"1x", "2x", "3"},
			want:    []string{"1x", "2x"},
		},
		{
			name:    "matchSubseq",
			matcher: matchSubseq,
			seed:    "xyz",
			src:     []string{"1xyz1", "2abc2", "123"},
			want:    []string{"1xyz1"},
		},
	}

	testRawFilterCandidates(t, tests)
}
