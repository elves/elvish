package eval

var builtinFnTests = []evalTest{
	{"nop", wantNothing},
	{"nop a b", wantNothing},
	{"nop &k=v", wantNothing},
	{"nop a b &k=v", wantNothing},

	{"kind-of bare 'str' [] [&] []{ }",
		want{out: strs("string", "string", "list", "map", "fn")}},

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
