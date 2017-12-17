package eval

func init() {
	addToEvalTests([]evalTest{
		{`path-base a/b/c.png`, want{out: strs("c.png")}},
		{`tilde-abbr $E:HOME/foobar`, want{out: strs("~/foobar")}},

		{`-is-dir ~/dir`, wantTrue}, // see testmain_test.go for setup
		{`-is-dir ~/lorem`, wantFalse},
	})
}
