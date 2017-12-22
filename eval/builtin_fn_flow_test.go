package eval

import "testing"

func TestBuiltinFnFlow(t *testing.T) {
	RunTests(t, libDir, []Test{
		{`run-parallel { put lorem } { echo ipsum }`,
			want{out: strs("lorem"), bytesOut: []byte("ipsum\n")}},

		{`put 1 233 | each put`, want{out: strs("1", "233")}},
		{`echo "1\n233" | each put`, want{out: strs("1", "233")}},
		{`each put [1 233]`, want{out: strs("1", "233")}},
		{`range 10 | each [x]{ if (== $x 4) { break }; put $x }`,
			want{out: strs("0", "1", "2", "3")}},
		{`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`,
			want{out: strs("0", "1", "2", "3"), err: errAny}},
		// TODO: test peach

		{`fail haha`, want{err: errAny}},
		{`return`, want{err: Return}},
	})
}
