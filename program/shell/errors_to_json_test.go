package shell

import (
	"errors"
	"testing"

	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

var errorsJSONTests = []struct {
	in   error
	want string
}{
	{&eval.CompilationError{Message: "This is a test error", Context: diag.SourceRange{Name: "CompileError", Source: "compile", Begin: 5, End: 7}}, `[{"fileName":"CompileError","start":5,"end":7,"message":"This is a test error"}]`},
	{&parse.MultiError{Entries: []*parse.Error{
		{Message: "This is a test error", Context: diag.SourceRange{Name: "ParseError", Source: "parse", Begin: 5, End: 7}},
		{Message: "This is a test error", Context: diag.SourceRange{Name: "ParseError", Source: "parse", Begin: 15, End: 16}},
	}}, `[{"fileName":"ParseError","start":5,"end":7,"message":"This is a test error"},{"fileName":"ParseError","start":15,"end":16,"message":"This is a test error"}]`},
	{errors.New("This is a test error"), `[{"message":"This is a test error"}]`},
}

func TestErrorsToJSON(t *testing.T) {
	for i, test := range errorsJSONTests {
		s := ErrorToJSON(test.in)
		if s != test.want {
			t.Errorf("test #%d (error = %s, json = %s) failed", i, test.in, s)
		}
	}
}
