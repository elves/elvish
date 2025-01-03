package eval_test

import (
	"testing"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/diag"
	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

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

// TODO: Turn this into a fuzz test.
func TestPartialCompilationError(t *testing.T) {
	for _, code := range transcriptCodes {
		testPartialError(t, code, func(src parse.Source) []*CompilationError {
			_, _, err := NewEvaler().Check(src, nil)
			return UnpackCompilationErrors(err)
		})
	}
}

// TODO: Deduplicate this against a similar helper in pkg/parse/fuzz_test.go
func testPartialError[T diag.ErrorTag](t *testing.T, full string, fn func(src parse.Source) []*diag.Error[T]) {
	if !utf8.ValidString(full) {
		t.Skip("not valid UTF-8")
	}
	errs := fn(parse.Source{Name: "fuzz.elv", Code: full})
	if len(errs) > 0 {
		t.Skip("code itself has error")
	}
	// If code has no error, then every prefix of it (as long as it's valid
	// UTF-8) should have either no errors or only partial errors.
	for i := range full {
		if i == 0 {
			continue
		}
		prefix := full[:i]
		errs := fn(parse.Source{Name: "fuzz.elv", Code: prefix})
		for _, err := range errs {
			if !err.Partial {
				t.Errorf("\n%s\n===========\nnon-partial error: %v\nfull code:\n===========\n%s\n", prefix, err, full)
			}
		}
	}
}
