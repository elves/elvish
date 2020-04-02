package parse

import (
	"testing"

	"github.com/elves/elvish/pkg/diag"
)

var multiErrorTests = []struct {
	multiError *MultiError
	indent     string
	wantError  string
	wantShow   string
}{
	{makeMultiError(), "", "no parse error", "no parse error"},
	// TODO: Add more complex test cases.
}

func TestMultiError(t *testing.T) {
	for _, test := range multiErrorTests {
		gotError := test.multiError.Error()
		if gotError != test.wantError {
			t.Errorf("got error %q, want %q", gotError, test.wantError)
		}
		gotShow := test.multiError.Show(test.indent)
		if gotShow != test.wantShow {
			t.Errorf("got show %q, want %q", gotShow, test.wantShow)
		}
	}
}

func makeMultiError(entries ...*diag.Error) *MultiError {
	return &MultiError{entries}
}
