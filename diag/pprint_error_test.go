package diag

import (
	"errors"
	"os"
	"strings"
	"testing"
)

var stderrBuf *strings.Builder

func setup() {
	stderrBuf = &strings.Builder{}
	stderr = stderrBuf
}

func teardown() {
	stderr = os.Stderr
}

type pprinterError struct{}

func (pprinterError) Error() string { return "error" }

func (pprinterError) PPrint(_ string) string { return "pprint" }

var pprintErrorTests = []struct {
	name    string
	err     error
	wantBuf string
}{
	{"A PPrinter error", pprinterError{}, "pprint\n"},
	{"A errors.New error", errors.New("ERROR"), "\033[31;1mERROR\033[m\n"},
}

func TestPPrintError(t *testing.T) {
	for _, test := range pprintErrorTests {
		t.Run(test.name, func(t *testing.T) {
			setup()
			defer teardown()
			PPrintError(test.err)
			if stderrBuf.String() != test.wantBuf {
				t.Errorf("Wrote %q, want %q", stderrBuf.String(), test.wantBuf)
			}
		})
	}
}
