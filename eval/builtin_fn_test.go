package eval

var builtinFnTests = []evalTest{
	// Builtin functions
	// -----------------

	{"kind-of bare 'str' [] [&] []{ }",
		want{out: strs("string", "string", "list", "map", "fn")}},

	{`put foo bar`, want{out: strs("foo", "bar")}},
	{`explode [foo bar]`, want{out: strs("foo", "bar")}},

	{`print [foo bar]`, want{bytesOut: []byte("[foo bar]")}},
	{`echo [foo bar]`, want{bytesOut: []byte("[foo bar]\n")}},
	{`pprint [foo bar]`, want{bytesOut: []byte("[\n foo\n bar\n]\n")}},

	{`print "a\nb" | slurp`, want{out: strs("a\nb")}},
	{`print "a\nb" | from-lines`, want{out: strs("a", "b")}},
	{`print "a\nb\n" | from-lines`, want{out: strs("a", "b")}},
	{`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`,
		want{out: []Value{
			ConvertToMap(map[Value]Value{
				String("k"): String("v"),
				String("a"): NewList(strs("1", "2")...)}),
			String("foo"),
		}}},
	{`echo 'invalid' | from-json`, want{err: errAny}},

	{`put "l\norem" ipsum | to-lines`,
		want{bytesOut: []byte("l\norem\nipsum\n")}},
	{`put [&k=v &a=[1 2]] foo | to-json`,
		want{bytesOut: []byte(`{"a":["1","2"],"k":"v"}
"foo"
`)}},

	{`joins : [/usr /bin /tmp]`, want{out: strs("/usr:/bin:/tmp")}},
	{`splits : /usr:/bin:/tmp`, want{out: strs("/usr", "/bin", "/tmp")}},
	{`replaces : / ":usr:bin:tmp"`, want{out: strs("/usr/bin/tmp")}},
	{`replaces &max=2 : / :usr:bin:tmp`, want{out: strs("/usr/bin:tmp")}},
	{`has-prefix golang go`, want{out: bools(true)}},
	{`has-prefix golang x`, want{out: bools(false)}},
	{`has-suffix golang x`, want{out: bools(false)}},

	{`keys [&]`, wantNothing},
	{`keys [&a=foo]`, want{out: strs("a")}},
	// Windows does not have an external sort command. Disabled until we have a
	// builtin sort command.
	// {`keys [&a=foo &b=bar] | each echo | sort | each put`, want{out: strs("a", "b")}},

	{`==s haha haha`, want{out: bools(true)}},
	{`==s 10 10.0`, want{out: bools(false)}},
	{`<s a b`, want{out: bools(true)}},
	{`<s 2 10`, want{out: bools(false)}},

	{`run-parallel { put lorem } { echo ipsum }`,
		want{out: strs("lorem"), bytesOut: []byte("ipsum\n")}},

	{`fail haha`, want{err: errAny}},
	{`return`, want{err: Return}},

	{`f=(constantly foo); $f; $f`, want{out: strs("foo", "foo")}},
	{`(constantly foo) bad`, want{err: errAny}},
	{`put 1 233 | each put`, want{out: strs("1", "233")}},
	{`echo "1\n233" | each put`, want{out: strs("1", "233")}},
	{`each put [1 233]`, want{out: strs("1", "233")}},
	{`range 10 | each [x]{ if (== $x 4) { break }; put $x }`,
		want{out: strs("0", "1", "2", "3")}},
	{`range 10 | each [x]{ if (== $x 4) { fail haha }; put $x }`,
		want{out: strs("0", "1", "2", "3"), err: errAny}},
	{`repeat 4 foo`, want{out: strs("foo", "foo", "foo", "foo")}},
	// TODO: test peach

	{`range 3`, want{out: strs("0", "1", "2")}},
	{`range 1 3`, want{out: strs("1", "2")}},
	{`range 0 10 &step=3`, want{out: strs("0", "3", "6", "9")}},
	{`put foo bar | all`, want{out: strs("foo", "bar")}},
	{`echo foobar | all`, want{bytesOut: []byte("foobar\n")}},
	{`{ put foo bar; echo foobar } | all`,
		want{out: strs("foo", "bar"), bytesOut: []byte("foobar\n")}},
	{`range 100 | take 2`, want{out: strs("0", "1")}},
	{`range 100 | drop 98`, want{out: strs("98", "99")}},
	{`range 100 | count`, want{out: strs("100")}},
	{`count [(range 100)]`, want{out: strs("100")}},

	{`echo "  ax  by cz  \n11\t22 33" | eawk [@a]{ put $a[-1] }`,
		want{out: strs("cz", "33")}},

	{`path-base a/b/c.png`, want{out: strs("c.png")}},

	// TODO test more edge cases
	{"+ 233100 233", want{out: strs("233333")}},
	{"- 233333 233100", want{out: strs("233")}},
	{"- 233", want{out: strs("-233")}},
	{"* 353 661", want{out: strs("233333")}},
	{"/ 233333 353", want{out: strs("661")}},
	{"/ 1 0", want{out: strs("+Inf")}},
	{"^ 16 2", want{out: strs("256")}},
	{"% 23 7", want{out: strs("2")}},

	{`== 1 1.0`, want{out: bools(true)}},
	{`== 10 0xa`, want{out: bools(true)}},
	{`== a a`, want{err: errAny}},
	{`> 0x10 1`, want{out: bools(true)}},

	{`is 1 1`, want{out: bools(true)}},
	{`is [] []`, want{out: bools(true)}},
	{`is [1] [1]`, want{out: bools(false)}},
	{`eq 1 1`, want{out: bools(true)}},
	{`eq [] []`, want{out: bools(true)}},

	{`ord a`, want{out: strs("0x61")}},
	{`base 16 42 233`, want{out: strs("2a", "e9")}},
	{`wcswidth 你好`, want{out: strs("4")}},
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

	{`put (assoc [0] 0 zero)[0]`, want{out: strs("zero")}},
	{`put (assoc [&] k v)[k]`, want{out: strs("v")}},
	{`put (assoc [&k=v] k v2)[k]`, want{out: strs("v2")}},
	{`has-key (dissoc [&k=v] k) k`, want{out: bools(false)}},
}

func init() {
	addToEvalTests(builtinFnTests)
}
