package eval

var builtinFnTests = []evalTest{
	// Builtin functions
	// -----------------

	{"kind-of bare 'str' [] [&] []{ }",
		want{out: strs("string", "string", "list", "map", "fn")}},

	{`is 1 1`, wantTrue},
	{`is [] []`, wantTrue},
	{`is [1] [1]`, wantFalse},
	{`eq 1 1`, wantTrue},
	{`eq [] []`, wantTrue},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
}

func init() {
	addToEvalTests(builtinFnTests)
}
