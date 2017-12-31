package eval

import (
	"testing"

	"github.com/elves/elvish/eval/types"
)

func TestBuiltinFnIO(t *testing.T) {
	runTests(t, []Test{
		{`put foo bar`, want{out: strs("foo", "bar")}},

		{`print [foo bar]`, want{bytesOut: []byte("[foo bar]")}},
		{`echo [foo bar]`, want{bytesOut: []byte("[foo bar]\n")}},
		{`pprint [foo bar]`, want{bytesOut: []byte("[\n foo\n bar\n]\n")}},
		NewTest(`repr foo bar ['foo bar']`).WantBytesOutString("foo bar ['foo bar']\n"),

		{`print "a\nb" | slurp`, want{out: strs("a\nb")}},
		{`print "a\nb" | from-lines`, want{out: strs("a", "b")}},
		{`print "a\nb\n" | from-lines`, want{out: strs("a", "b")}},
		{`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`,
			want{out: []types.Value{
				types.MakeMap(map[types.Value]types.Value{
					String("k"): String("v"),
					String("a"): types.MakeList(strs("1", "2")...)}),
				String("foo"),
			}}},
		{`echo 'invalid' | from-json`, want{err: errAny}},

		{`put "l\norem" ipsum | to-lines`,
			want{bytesOut: []byte("l\norem\nipsum\n")}},
		{`put [&k=v &a=[1 2]] foo | to-json`,
			want{bytesOut: []byte(`{"a":["1","2"],"k":"v"}
"foo"
`)}},
	})
}
