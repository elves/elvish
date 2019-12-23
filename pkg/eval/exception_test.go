package eval

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

func TestCause(t *testing.T) {
	err := errors.New("ordinary error")
	exc := &Exception{Cause: err}
	tt.Test(t, tt.Fn("Cause", Cause), tt.Table{
		tt.Args(err).Rets(err),
		tt.Args(exc).Rets(err),
	})
}

func TestException(t *testing.T) {
	Test(t,
		That("kind-of ?(fail foo)").Puts("exception"),
	)
}
