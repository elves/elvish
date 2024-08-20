package vals

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/tt"
)

type customAssocer struct{}

var errCustomAssoc = errors.New("custom assoc error")

func (a customAssocer) Assoc(k, v any) (any, error) {
	return "custom result", errCustomAssoc
}

func TestAssoc(t *testing.T) {
	tt.Test(t, Assoc,
		// String
		Args("0123", "0", "foo").Rets("foo123", nil),
		Args("0123", "1..3", "bar").Rets("0bar3", nil),
		Args("0123", "1..3", 12).Rets(nil, errReplacementMustBeString),
		Args("0123", "x", "y").Rets(nil, errIndexMustBeInteger),

		// List
		Args(MakeList("0", "1", "2", "3"), "0", "foo").Rets(
			eq(MakeList("foo", "1", "2", "3")), nil),
		Args(MakeList("0", "1", "2", "3"), 0, "foo").Rets(
			eq(MakeList("foo", "1", "2", "3")), nil),
		Args(MakeList("0"), MakeList("0"), "1").Rets(nil, errIndexMustBeInteger),
		Args(MakeList("0"), "1", "x").Rets(nil, errs.OutOfRange{
			What: "index", ValidLow: "0", ValidHigh: "0", Actual: "1"}),
		// TODO: Support list assoc with slice
		Args(MakeList("0", "1", "2", "3"), "1..3", MakeList("foo")).Rets(
			nil, errAssocWithSlice),

		// Map
		Args(MakeMap("k", "v", "k2", "v2"), "k", "newv").Rets(
			eq(MakeMap("k", "newv", "k2", "v2")), nil),
		Args(MakeMap("k", "v"), "k2", "v2").Rets(
			eq(MakeMap("k", "v", "k2", "v2")), nil),

		// Field map
		Args(fieldMap{"lorem", "ipsum", 23}, "foo-bar", "x").Rets(
			eq(MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", "x")), nil),

		Args(customAssocer{}, "x", "y").Rets("custom result", errCustomAssoc),

		Args(struct{}{}, "x", "y").Rets(nil, errAssocUnsupported),
	)
}
