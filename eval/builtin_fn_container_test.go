package eval

func init() {
	addToEvalTests([]evalTest{
		{`range 3`, want{out: strs("0", "1", "2")}},
		{`range 1 3`, want{out: strs("1", "2")}},
		{`range 0 10 &step=3`, want{out: strs("0", "3", "6", "9")}},
		{`repeat 4 foo`, want{out: strs("foo", "foo", "foo", "foo")}},
		{`explode [foo bar]`, want{out: strs("foo", "bar")}},

		{`put (assoc [0] 0 zero)[0]`, want{out: strs("zero")}},
		{`put (assoc [&] k v)[k]`, want{out: strs("v")}},
		{`put (assoc [&k=v] k v2)[k]`, want{out: strs("v2")}},
		{`has-key (dissoc [&k=v] k) k`, want{out: bools(false)}},

		{`put foo bar | all`, want{out: strs("foo", "bar")}},
		{`echo foobar | all`, want{bytesOut: []byte("foobar\n")}},
		{`{ put foo bar; echo foobar } | all`,
			want{out: strs("foo", "bar"), bytesOut: []byte("foobar\n")}},
		{`range 100 | take 2`, want{out: strs("0", "1")}},
		{`range 100 | drop 98`, want{out: strs("98", "99")}},

		{`has-key [foo bar] 0`, want{out: bools(true)}},
		{`has-key [foo bar] 0:1`, want{out: bools(true)}},
		{`has-key [foo bar] 0:20`, want{out: bools(false)}},
		{`has-key [&lorem=ipsum &foo=bar] lorem`, want{out: bools(true)}},
		{`has-key [&lorem=ipsum &foo=bar] loremwsq`, want{out: bools(false)}},
		{`has-value [&lorem=ipsum &foo=bar] lorem`, want{out: bools(false)}},
		{`has-value [&lorem=ipsum &foo=bar] bar`, want{out: bools(true)}},
		{`has-value [foo bar] bar`, want{out: bools(true)}},
		{`has-value [foo bar] badehose`, want{out: bools(false)}},
		{`has-value "foo" o`, want{out: bools(true)}},
		{`has-value "foo" d`, want{out: bools(false)}},

		{`range 100 | count`, want{out: strs("100")}},
		{`count [(range 100)]`, want{out: strs("100")}},

		{`keys [&]`, wantNothing},
		{`keys [&a=foo]`, want{out: strs("a")}},
		// Windows does not have an external sort command. Disabled until we have a
		// builtin sort command.
		// {`keys [&a=foo &b=bar] | each echo | sort | each put`, want{out: strs("a", "b")}},
	})
}
