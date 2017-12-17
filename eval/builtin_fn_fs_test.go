package eval

func init() {
	addToEvalTests([]evalTest{
		{`path-base a/b/c.png`, want{out: strs("c.png")}},
	})
}
