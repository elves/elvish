package eval

import "testing"

var builtinFnTests = []Test{
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
	{`not ?(fail x)`, wantTrue},
	{`not $true`, wantFalse},
	{`not 0`, wantFalse},

	{`is 1 1`, wantTrue},
	{`is a b`, wantFalse},
	{`is [] []`, wantTrue},
	{`is [1] [1]`, wantFalse},
	{`eq 1 1`, wantTrue},
	{`eq a b`, wantFalse},
	{`eq [] []`, wantTrue},
	{`eq [1] [1]`, wantTrue},
	{`not-eq a b`, wantTrue},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
}

func TestBuiltinFn(t *testing.T) {
	RunTests(t, libDir, builtinFnTests)
}
