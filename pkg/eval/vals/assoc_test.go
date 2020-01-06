package vals

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/tt"
)

type customAssocer struct{}

var errCustomAssoc = errors.New("custom assoc error")

func (a customAssocer) Assoc(k, v interface{}) (interface{}, error) {
	return "custom result", errCustomAssoc
}

var assocTests = tt.Table{
	Args("0123", "0", "foo").Rets("foo123", nil),
	Args("0123", "1:3", "bar").Rets("0bar3", nil),
	Args("0123", "1:3", 12).Rets(nil, errReplacementMustBeString),
	Args("0123", "x", "y").Rets(nil, errIndexMustBeInteger),

	Args(MakeList("0", "1", "2", "3"), "0", "foo").Rets(
		eq(MakeList("foo", "1", "2", "3")), nil),
	Args(MakeList("0", "1", "2", "3"), 0.0, "foo").Rets(
		eq(MakeList("foo", "1", "2", "3")), nil),
	Args(MakeList("0"), MakeList("0"), "1").Rets(nil, errIndexMustBeInteger),
	Args(MakeList("0"), "1", "x").Rets(nil, errIndexOutOfRange),
	// TODO: Support list assoc with slice
	Args(MakeList("0", "1", "2", "3"), "1:3", MakeList("foo")).Rets(
		nil, errAssocWithSlice),

	Args(MakeMap("k", "v", "k2", "v2"), "k", "newv").Rets(
		eq(MakeMap("k", "newv", "k2", "v2")), nil),
	Args(MakeMap("k", "v"), "k2", "v2").Rets(
		eq(MakeMap("k", "v", "k2", "v2")), nil),

	Args(testStructMap{"foo", 1.0}, "name", "bar").
		Rets(testStructMap{"bar", 1.0}, nil),
	Args(testStructMap{"foo", 1.0}, "score-number", "2.0").
		Rets(testStructMap{"foo", 2.0}, nil),
	Args(testStructMap{"foo", 1.0}, "score-number", "bad number").
		Rets(nil, cannotParseAs{"number", `'bad number'`}),

	Args(customAssocer{}, "x", "y").Rets("custom result", errCustomAssoc),

	Args(struct{}{}, "x", "y").Rets(nil, errAssocUnsupported),
}

func TestAssoc(t *testing.T) {
	tt.Test(t, tt.Fn("Assoc", Assoc), assocTests)
}
