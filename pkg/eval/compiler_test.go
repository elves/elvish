package eval_test

import (
	"bytes"
	"strings"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
)

func TestDeprecatedBuiltin(t *testing.T) {
	testCompileTimeDeprecation(t, "fopen a", `the "fopen" command is deprecated`, 16)
	// Deprecations of other builtins are implemented in the same way, so we
	// don't test them repeatedly
}

func testCompileTimeDeprecation(t *testing.T, code, wantWarning string, level int) {
	t.Helper()
	restore := prog.SetDeprecationLevel(level)
	defer restore()

	ev := NewEvaler()
	errOutput := new(bytes.Buffer)

	parseErr, compileErr := ev.Check(parse.Source{Code: code}, errOutput)
	if parseErr != nil {
		t.Errorf("got parse err %v", parseErr)
	}
	if compileErr != nil {
		t.Errorf("got compile err %v", compileErr)
	}

	warning := errOutput.String()
	if !strings.Contains(warning, wantWarning) {
		t.Errorf("got warning %q, want warning containing %q", warning, wantWarning)
	}
}
