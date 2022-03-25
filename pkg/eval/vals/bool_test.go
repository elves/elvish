package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type customBooler struct{ b bool }

func (b customBooler) Bool() bool { return b.b }

type customNonBooler struct{}

func TestBool(t *testing.T) {
	tt.Test(t, tt.Fn("Bool", Bool), tt.Table{
		tt.Args(nil).Rets(false),

		tt.Args(true).Rets(true),
		tt.Args(false).Rets(false),

		tt.Args(customBooler{true}).Rets(true),
		tt.Args(customBooler{false}).Rets(false),

		tt.Args(customNonBooler{}).Rets(true),
	})
}
