package errs

import (
	"testing"
)

var errorMessageTests = []struct {
	err     error
	wantMsg string
}{
	{
		OutOfRange{What: "list index here", ValidLow: 0, ValidHigh: 2, Actual: "3"},
		"out of range: list index here must be from 0 to 2, but is 3",
	},
	{
		OutOfRange{What: "list index here", ValidLow: 1, ValidHigh: 0, Actual: "0"},
		"out of range: list index here has no valid value, but is 0",
	},
	{
		ArityMismatch{What: "arguments here", ValidLow: 2, ValidHigh: 2, Actual: 3},
		"arity mismatch: arguments here must be 2 values, but is 3 values",
	},
	{
		ArityMismatch{What: "arguments here", ValidLow: 2, ValidHigh: -1, Actual: 1},
		"arity mismatch: arguments here must be 2 or more values, but is 1 value",
	},
	{
		ArityMismatch{What: "arguments here", ValidLow: 2, ValidHigh: 3, Actual: 1},
		"arity mismatch: arguments here must be 2 to 3 values, but is 1 value",
	},
}

func TestErrorMessages(t *testing.T) {
	for _, test := range errorMessageTests {
		if gotMsg := test.err.Error(); gotMsg != test.wantMsg {
			t.Errorf("got message %v, want %v", gotMsg, test.wantMsg)
		}
	}
}
