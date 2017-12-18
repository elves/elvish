package eval

func init() {
	addToEvalTests([]evalTest{
		{`== 1 1.0`, wantTrue},
		{`== 10 0xa`, wantTrue},
		{`== a a`, want{err: errAny}},
		{`> 0x10 1`, wantTrue},

		// TODO test more edge cases
		{"+ 233100 233", want{out: strs("233333")}},
		{"- 233333 233100", want{out: strs("233")}},
		{"- 233", want{out: strs("-233")}},
		{"* 353 661", want{out: strs("233333")}},
		{"/ 233333 353", want{out: strs("661")}},
		{"/ 1 0", want{out: strs("+Inf")}},
		{"^ 16 2", want{out: strs("256")}},
		{"% 23 7", want{out: strs("2")}},
	})
}
