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

	{`bool $true`, wantTrue},
	{`bool a`, wantTrue},
	{`bool [a]`, wantTrue},
	// "Empty" values are also true in Elvish
	{`bool []`, wantTrue},
	{`bool [&]`, wantTrue},
	{`bool 0`, wantTrue},
	{`bool ""`, wantTrue},
	// Only errors and $false are false
	{`bool ?(fail x)`, wantFalse},
	{`bool $false`, wantFalse},

	{`not $false`, wantTrue},
	{`not $true`, wantFalse},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
}

func init() {
	addToEvalTests(builtinFnTests)
}
