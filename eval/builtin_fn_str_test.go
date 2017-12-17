package eval

var ()

func init() {
	addToEvalTests([]evalTest{
		{`==s haha haha`, wantTrue},
		{`==s 10 10.0`, wantFalse},
		{`<s a b`, wantTrue},
		{`<s 2 10`, wantFalse},

		{`joins : [/usr /bin /tmp]`, want{out: strs("/usr:/bin:/tmp")}},
		{`splits : /usr:/bin:/tmp`, want{out: strs("/usr", "/bin", "/tmp")}},
		{`replaces : / ":usr:bin:tmp"`, want{out: strs("/usr/bin/tmp")}},
		{`replaces &max=2 : / :usr:bin:tmp`, want{out: strs("/usr/bin:tmp")}},

		{`ord a`, want{out: strs("0x61")}},
		{`base 16 42 233`, want{out: strs("2a", "e9")}},
		{`wcswidth 你好`, want{out: strs("4")}},

		{`has-prefix golang go`, wantTrue},
		{`has-prefix golang x`, wantFalse},
		{`has-suffix golang x`, wantFalse},

		{`echo "  ax  by cz  \n11\t22 33" | eawk [@a]{ put $a[-1] }`,
			want{out: strs("cz", "33")}},
	})
}
