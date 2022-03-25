package vals

import (
	"errors"
	"math/big"
	"testing"

	"src.elv.sh/pkg/tt"
)

// An implementation for Concatter that accepts strings, returns a special
// error when rhs is a float64, and returns ErrConcatNotImplemented when rhs is
// of other types.
type concatter struct{}

var errBadFloat64 = errors.New("float64 is bad")

func (concatter) Concat(rhs any) (any, error) {
	switch rhs := rhs.(type) {
	case string:
		return "concatter " + rhs, nil
	case float64:
		return nil, errBadFloat64
	default:
		return nil, ErrConcatNotImplemented
	}
}

// An implementation of RConcatter that accepts all types.
type rconcatter struct{}

func (rconcatter) RConcat(lhs any) (any, error) {
	return "rconcatter", nil
}

func TestConcat(t *testing.T) {
	tt.Test(t, tt.Fn("Concat", Concat), tt.Table{
		tt.Args("foo", "bar").Rets("foobar", nil),
		// string+number
		tt.Args("foo", 2).Rets("foo2", nil),
		tt.Args("foo", bigInt(z)).Rets("foo"+z, nil),
		tt.Args("foo", big.NewRat(1, 2)).Rets("foo1/2", nil),
		tt.Args("foo", 2.0).Rets("foo2.0", nil),
		// number+string
		tt.Args(2, "foo").Rets("2foo", nil),
		tt.Args(bigInt(z), "foo").Rets(z+"foo", nil),
		tt.Args(big.NewRat(1, 2), "foo").Rets("1/2foo", nil),
		tt.Args(2.0, "foo").Rets("2.0foo", nil),

		// LHS implements Concatter and succeeds
		tt.Args(concatter{}, "bar").Rets("concatter bar", nil),
		// LHS implements Concatter but returns ErrConcatNotImplemented; RHS
		// does not implement RConcatter
		tt.Args(concatter{}, 12).Rets(nil, cannotConcat{"!!vals.concatter", "number"}),
		// LHS implements Concatter but returns another error
		tt.Args(concatter{}, 12.0).Rets(nil, errBadFloat64),

		// LHS does not implement Concatter but RHS implements RConcatter
		tt.Args(12, rconcatter{}).Rets("rconcatter", nil),
	})
}
