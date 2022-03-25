package parse

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/tt"
)

var Args = tt.Args

func TestGetError(t *testing.T) {
	parseError := makeError()
	tt.Test(t, tt.Fn("GetError", GetError), tt.Table{
		Args(parseError).Rets(parseError),
		Args(errors.New("random error")).Rets((*Error)(nil)),
	})
}

var errorTests = []struct {
	err       *Error
	indent    string
	wantError string
	wantShow  string
}{
	{makeError(), "", "no parse error", "no parse error"},
	// TODO: Add more complex test cases.
}

func TestError(t *testing.T) {
	for _, test := range errorTests {
		gotError := test.err.Error()
		if gotError != test.wantError {
			t.Errorf("got error %q, want %q", gotError, test.wantError)
		}
		gotShow := test.err.Show(test.indent)
		if gotShow != test.wantShow {
			t.Errorf("got show %q, want %q", gotShow, test.wantShow)
		}
	}
}

func makeError(entries ...*diag.Error) *Error {
	return &Error{entries}
}
