package eval

import "testing"

func TestException(t *testing.T) {
	runTests(t, []Test{
		That("kind-of ?(fail foo)").Puts("exception"),
	})
}
