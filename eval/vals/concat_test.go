package vals

import (
	"errors"
	"testing"

	"github.com/elves/elvish/tt"
)

// An implementation for Concatter that accepts strings, returns a special
// error when rhs is a float64, and returns ErrConcatNotImplemented when rhs is
// of other types.
type concatter struct{}

func (concatter) Concat(rhs interface{}) (interface{}, error) {
	switch rhs := rhs.(type) {
	case string:
		return "concatter " + rhs, nil
	case float64:
		return nil, errors.New("float64 is bad")
	default:
		return nil, ErrConcatNotImplemented
	}
}

// An implementation of RConcatter that accepts all types.
type rconcatter struct{}

func (rconcatter) RConcat(lhs interface{}) (interface{}, error) {
	return "rconcatter", nil
}

func TestConcat(t *testing.T) {
	tt.Test(t, tt.Fn("Concat", Concat), tt.Table{
		Args("foo", "bar").Rets("foobar", nil),

		// LHS implements Concatter and succeeds
		Args(concatter{}, "bar").Rets("concatter bar", nil),
		// LHS implements Concatter but returns ErrConcatNotImplemented; RHS
		// does not implement RConcatter
		Args(concatter{}, 12).Rets(nil, errors.New("unsupported concat: !!vals.concatter and !!int")),
		// LHS implements Concatter but returns another error
		Args(concatter{}, 12.0).Rets(nil, errors.New("float64 is bad")),

		// LHS does not implement Concatter but RHS implements RConcatter
		Args(12, rconcatter{}).Rets("rconcatter", nil),
	})
}
