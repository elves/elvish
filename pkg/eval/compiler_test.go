package eval_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/testutil"
)

func TestDeprecatedBuiltin(t *testing.T) {
	testCompileTimeDeprecation(t, "eawk", `the "eawk" command is deprecated`, 20)

	// Deprecations of other builtins are implemented in the same way, so we
	// don't test them repeatedly
}

func testCompileTimeDeprecation(t *testing.T, code, wantWarning string, level int) {
	t.Helper()
	testutil.Set(t, &prog.DeprecationLevel, level)

	ev := NewEvaler()
	errOutput := new(bytes.Buffer)

	parseErr, _, compileErr := ev.Check(parse.Source{Name: "[test]", Code: code}, errOutput)
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

func TestMultipleCompileationErrors(t *testing.T) {
	Test(t,
		That("echo $x; echo $y").DoesNotCompile(
			"variable $x not found",
			"variable $y not found"),
	)
}

var autofixTests = []struct {
	Name string
	Code string

	WantAutofixes []string
}{
	{
		Name: "get variable from unimported builtin module",
		Code: "echo $mod1:foo",

		WantAutofixes: []string{"use mod1"},
	},
	{
		Name: "set variable from unimported builtin module",
		Code: "set mod1:foo = bar",

		WantAutofixes: []string{"use mod1"},
	},
	{
		Name: "tmp set variable from unimported builtin module",
		Code: "tmp mod1:foo = bar",

		WantAutofixes: []string{"use mod1"},
	},
	{
		Name: "call command from unimported builtin module",
		Code: "mod1:foo",

		WantAutofixes: []string{"use mod1"},
	},
	{
		Name: "no autofix for using variable from imported module",
		Code: "use mod1; echo $mod1:foo",

		WantAutofixes: nil,
	},
	{
		Name: "no autofix for using variable from non-builtin module",
		Code: "echo $mod-external:foo",

		WantAutofixes: nil,
	},
}

func TestAutofix(t *testing.T) {
	ev := NewEvaler()
	ev.AddModule("mod1", &Ns{})
	for _, tc := range autofixTests {
		t.Run(tc.Name, func(t *testing.T) {
			_, autofixes, _ := ev.Check(parse.Source{Name: "[test]", Code: tc.Code}, nil)
			if diff := cmp.Diff(tc.WantAutofixes, autofixes); diff != "" {
				t.Errorf("autofixes (-want +got):\n%s", diff)
			}
		})
	}
}
