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
	// We don't depend on user visible deprecations since they have significant churn and we don't
	// want to udpate this unit test when a deprecation is removed. Instead, create builtins and
	// related deprecations for purposes of this unit test.
	AddBuiltinFns(map[string]interface{}{
		"deprecated1": func() {},
		"deprecated2": func() {},
	})
	const dep1msg = "b:d1 is deprecated"
	const dep2msg = "b:d2 is deprecated"
	DeprecatedVars["builtin:deprecated1~"] = DeprecatedWhen{1, dep1msg}
	DeprecatedVars["builtin:deprecated2~"] = DeprecatedWhen{2, dep2msg}

	testCompileTimeDeprecation(t, "builtin:deprecated2", dep2msg, 99)
	testCompileTimeDeprecation(t, "deprecated2", dep2msg, 99)
	testCompileTimeDeprecation(t, "deprecated2", dep2msg, 2)
	// There should be no deprecation warning since we're asking about versions older than
	// when `deprecated2` was deprecated.
	testCompileTimeDeprecation(t, "deprecated2", ``, 1)

	testCompileTimeDeprecation(t, "builtin:deprecated1", dep1msg, 99)
	testCompileTimeDeprecation(t, "deprecated1", dep1msg, 99)
	testCompileTimeDeprecation(t, "deprecated1", dep1msg, 1)
	// There should be no deprecation warning since we're asking about versions older than
	// when `deprecated1` was deprecated.
	testCompileTimeDeprecation(t, "deprecated1", ``, 0)
}

func testCompileTimeDeprecation(t *testing.T, code, wantWarning string, level int) {
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
	if wantWarning == "" {
		if warning != "" {
			t.Errorf("got warning %q, want no warning", warning)
		}
	} else {
		if !strings.Contains(warning, wantWarning) {
			t.Errorf("got warning %q, want warning containing %q", warning, wantWarning)
		}
	}
}
