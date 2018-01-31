package types

import (
	"errors"
	"testing"

	"github.com/elves/elvish/tt"
)

type customAssocer struct{}

var customAssocError = errors.New("custom assoc error")

func (a customAssocer) Assoc(k, v interface{}) (interface{}, error) {
	return "custom result", customAssocError
}

var assocTests = tt.Table{
	Args("0123", "0", "foo").Rets("foo123", nil),
	Args("0123", "1:3", "bar").Rets("0bar3", nil),
	Args("0123", "1:3", 12).Rets(nil, errReplacementMustBeString),

	Args(MakeList("0", "1", "2", "3"), "0", "foo").Rets(
		eq(MakeList("foo", "1", "2", "3")), nil),
	// TODO: Support list assoc with slice
	Args(MakeList("0", "1", "2", "3"), "1:3", MakeList("foo")).Rets(
		nil, errAssocWithSlice),

	Args(MakeMapFromKV("k", "v", "k2", "v2"), "k", "newv").Rets(
		eq(MakeMapFromKV("k", "newv", "k2", "v2")), nil),
	Args(MakeMapFromKV("k", "v"), "k2", "v2").Rets(
		eq(MakeMapFromKV("k", "v", "k2", "v2")), nil),

	Args(customAssocer{}, "x", "y").Rets("custom result", customAssocError),
}

func TestAssoc(t *testing.T) {
	tt.Test(t, tt.Fn("Assoc", Assoc), assocTests)
}
