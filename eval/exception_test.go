package eval

import "testing"

func TestException(t *testing.T) {
	runTests(t, []Test{
		NewTest("kind-of ?(fail foo)").WantOutStrings("exception"),
	})
}
