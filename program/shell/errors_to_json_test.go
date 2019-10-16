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
	{
		&eval.CompilationError{
			Message: "ERR",
			Context: diag.SourceRange{Name: "file", Begin: 5, End: 7}},
		`[{"fileName":"file","start":5,"end":7,"message":"ERR"}]`,
	},
	{
		&parse.MultiError{Entries: []*parse.Error{
			{
				Message: "ERR1",
				Context: diag.SourceRange{Name: "file1", Begin: 5, End: 7}},
			{
				Message: "ERR2",
				Context: diag.SourceRange{Name: "file2", Begin: 15, End: 16}},
		}},
		`[{"fileName":"file1","start":5,"end":7,"message":"ERR1"},` +
			`{"fileName":"file2","start":15,"end":16,"message":"ERR2"}]`,
	},
	{
		errors.New("ERR"),
		`[{"message":"ERR"}]`,
	},
}

func TestErrorsToJSON(t *testing.T) {
	for i, test := range errorsJSONTests {
		s := ErrorToJSON(test.in)
		if s != test.want {
			t.Errorf("test #%d (error = %s, json = %s) failed", i, test.in, s)
		}
	}
}
