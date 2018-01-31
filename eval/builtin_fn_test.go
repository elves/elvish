package eval

import "testing"

var builtinFnTests = []Test{
	NewTest("kind-of $nop~").WantOutStrings("fn"),
	NewTest("eq $nop~ { }").WantOutBools(false),
	NewTest("put [&$nop~= foo][$nop~]").WantOutStrings("foo"),
	NewTest("repr $nop~").WantBytesOutString("<builtin nop>\n"),

	{"nop", wantNothing},
	{"nop a b", wantNothing},
	{"nop &k=v", wantNothing},
	{"nop a b &k=v", wantNothing},

	{"kind-of bare 'str' [] [&] []{ }",
		want{out: strs("string", "string", "list", "map", "fn")}},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
}

func TestBuiltinFn(t *testing.T) {
	runTests(t, builtinFnTests)
}
