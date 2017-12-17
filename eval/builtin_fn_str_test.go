package eval

func init() {
	addToEvalTests([]evalTest{
		{`==s haha haha`, want{out: bools(true)}},
		{`==s 10 10.0`, want{out: bools(false)}},
		{`<s a b`, want{out: bools(true)}},
		{`<s 2 10`, want{out: bools(false)}},

		{`joins : [/usr /bin /tmp]`, want{out: strs("/usr:/bin:/tmp")}},
		{`splits : /usr:/bin:/tmp`, want{out: strs("/usr", "/bin", "/tmp")}},
		{`replaces : / ":usr:bin:tmp"`, want{out: strs("/usr/bin/tmp")}},
		{`replaces &max=2 : / :usr:bin:tmp`, want{out: strs("/usr/bin:tmp")}},

		{`ord a`, want{out: strs("0x61")}},
		{`base 16 42 233`, want{out: strs("2a", "e9")}},
		{`wcswidth 你好`, want{out: strs("4")}},

		{`has-prefix golang go`, want{out: bools(true)}},
		{`has-prefix golang x`, want{out: bools(false)}},
		{`has-suffix golang x`, want{out: bools(false)}},

		{`echo "  ax  by cz  \n11\t22 33" | eawk [@a]{ put $a[-1] }`,
			want{out: strs("cz", "33")}},
	})
}
