package diag

import (
	"errors"
	"strings"
	"testing"
)

type showerError struct{}

func (showerError) Error() string { return "error" }

func (showerError) Show(_ string) string { return "show" }

var showErrorTests = []struct {
	name    string
	err     error
	wantBuf string
}{
	{"A Shower error", showerError{}, "show\n"},
	{"A errors.New error", errors.New("ERROR"), "\033[31;1mERROR\033[m\n"},
}

func TestShowError(t *testing.T) {
	for _, test := range showErrorTests {
		t.Run(test.name, func(t *testing.T) {
			sb := &strings.Builder{}
			ShowError(sb, test.err)
			if sb.String() != test.wantBuf {
				t.Errorf("Wrote %q, want %q", sb.String(), test.wantBuf)
			}
		})
	}
}
