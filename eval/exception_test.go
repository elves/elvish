package eval

import "testing"

func TestException(t *testing.T) {
	test(t, []TestCase{
		That("kind-of ?(fail foo)").Puts("exception"),
	})
}
