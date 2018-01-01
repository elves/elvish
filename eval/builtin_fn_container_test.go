package eval

import "testing"

func TestBuiltinFnContainer(t *testing.T) {
	runTests(t, []Test{
		{`range 3`, want{out: strs("0", "1", "2")}},
		{`range 1 3`, want{out: strs("1", "2")}},
		{`range 0 10 &step=3`, want{out: strs("0", "3", "6", "9")}},
		{`repeat 4 foo`, want{out: strs("foo", "foo", "foo", "foo")}},
		{`explode [foo bar]`, want{out: strs("foo", "bar")}},

		{`put (assoc [0] 0 zero)[0]`, want{out: strs("zero")}},
		{`put (assoc [&] k v)[k]`, want{out: strs("v")}},
		{`put (assoc [&k=v] k v2)[k]`, want{out: strs("v2")}},
		{`has-key (dissoc [&k=v] k) k`, wantFalse},

		{`put foo bar | all`, want{out: strs("foo", "bar")}},
		{`echo foobar | all`, want{bytesOut: []byte("foobar\n")}},
		{`{ put foo bar; echo foobar } | all`,
			want{out: strs("foo", "bar"), bytesOut: []byte("foobar\n")}},
		{`range 100 | take 2`, want{out: strs("0", "1")}},
		{`range 100 | drop 98`, want{out: strs("98", "99")}},

		{`has-key [foo bar] 0`, wantTrue},
		{`has-key [foo bar] 0:1`, wantTrue},
		{`has-key [foo bar] 0:20`, wantFalse},
		{`has-key [&lorem=ipsum &foo=bar] lorem`, wantTrue},
		{`has-key [&lorem=ipsum &foo=bar] loremwsq`, wantFalse},
		{`has-value [&lorem=ipsum &foo=bar] lorem`, wantFalse},
		{`has-value [&lorem=ipsum &foo=bar] bar`, wantTrue},
		{`has-value [foo bar] bar`, wantTrue},
		{`has-value [foo bar] badehose`, wantFalse},
		{`has-value "foo" o`, wantTrue},
		{`has-value "foo" d`, wantFalse},

		{`range 100 | count`, want{out: strs("100")}},
		{`count [(range 100)]`, want{out: strs("100")}},
		{`count 1 2 3`, want{err: errAny}},

		{`keys [&]`, wantNothing},
		{`keys [&a=foo]`, want{out: strs("a")}},
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		// {`keys [&a=foo &b=bar] | each echo | sort | each $put~`, want{out: strs("a", "b")}},
	})
}
