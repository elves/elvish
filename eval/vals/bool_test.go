package vals

import (
	"testing"

	"github.com/elves/elvish/tt"
)

type customBooler struct{ b bool }

func (b customBooler) Bool() bool { return b.b }

type customNonBooler struct{}

var boolTests = tt.Table{
	Args(true).Rets(true),
	Args(false).Rets(false),

	Args(customBooler{true}).Rets(true),
	Args(customBooler{false}).Rets(false),

	Args(customNonBooler{}).Rets(true),
}

func TestBool(t *testing.T) {
	tt.Test(t, tt.Fn("Bool", Bool), boolTests)
}
