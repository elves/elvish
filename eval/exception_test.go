package eval

import "testing"

func TestException(t *testing.T) {
	Test(t, []TestCase{
		That("kind-of ?(fail foo)").Puts("exception"),
	})
}
