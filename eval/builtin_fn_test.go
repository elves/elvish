package eval

var builtinFnTests = []evalTest{
	// Builtin functions
	// -----------------

	{"kind-of bare 'str' [] [&] []{ }",
		want{out: strs("string", "string", "list", "map", "fn")}},

	{`is 1 1`, want{out: bools(true)}},
	{`is [] []`, want{out: bools(true)}},
	{`is [1] [1]`, want{out: bools(false)}},
	{`eq 1 1`, want{out: bools(true)}},
	{`eq [] []`, want{out: bools(true)}},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
}

func init() {
	addToEvalTests(builtinFnTests)
}
