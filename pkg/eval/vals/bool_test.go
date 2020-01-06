package vals

import (
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

type customBooler struct{ b bool }

func (b customBooler) Bool() bool { return b.b }

type customNonBooler struct{}

func TestBool(t *testing.T) {
	Test(t, Fn("Bool", Bool), Table{
		Args(nil).Rets(false),

		Args(true).Rets(true),
		Args(false).Rets(false),

		Args(customBooler{true}).Rets(true),
		Args(customBooler{false}).Rets(false),

		Args(customNonBooler{}).Rets(true),
	})
}
