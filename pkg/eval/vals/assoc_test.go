package vals

import (
	"errors"
	"testing"

	. "github.com/elves/elvish/pkg/tt"
)

type customAssocer struct{}

var errCustomAssoc = errors.New("custom assoc error")

func (a customAssocer) Assoc(k, v interface{}) (interface{}, error) {
	return "custom result", errCustomAssoc
}

func TestAssoc(t *testing.T) {
	Test(t, Fn("Assoc", Assoc), Table{
		Args("0123", "0", "foo").Rets("foo123", nil),
		Args("0123", "1:3", "bar").Rets("0bar3", nil),
		Args("0123", "1:3", 12).Rets(nil, errReplacementMustBeString),
		Args("0123", "x", "y").Rets(nil, errIndexMustBeInteger),

		Args(MakeList("0", "1", "2", "3"), "0", "foo").Rets(
			Eq(MakeList("foo", "1", "2", "3")), nil),
		Args(MakeList("0", "1", "2", "3"), 0.0, "foo").Rets(
			Eq(MakeList("foo", "1", "2", "3")), nil),
		Args(MakeList("0"), MakeList("0"), "1").Rets(nil, errIndexMustBeInteger),
		Args(MakeList("0"), "1", "x").Rets(nil, ErrIndexOutOfRange),
		// TODO: Support list assoc with slice
		Args(MakeList("0", "1", "2", "3"), "1:3", MakeList("foo")).Rets(
			nil, errAssocWithSlice),

		Args(MakeMap("k", "v", "k2", "v2"), "k", "newv").Rets(
			Eq(MakeMap("k", "newv", "k2", "v2")), nil),
		Args(MakeMap("k", "v"), "k2", "v2").Rets(
			Eq(MakeMap("k", "v", "k2", "v2")), nil),

		Args(customAssocer{}, "x", "y").Rets("custom result", errCustomAssoc),

		Args(struct{}{}, "x", "y").Rets(nil, errAssocUnsupported),
	})
}
