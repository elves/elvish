package shell

import (
	"errors"
	"testing"

	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/tt"
)

func TestErrorsToJSON(t *testing.T) {
	tt.Test(t, tt.Fn("errorToJSON", errorToJSON), tt.Table{
		tt.Args(eval.NewCompilationError(
			"ERR",
			diag.Context{Name: "file", Begin: 5, End: 7}),
		).Rets(
			[]byte(`[{"fileName":"file","start":5,"end":7,"message":"ERR"}]`),
		),
		tt.Args(&parse.MultiError{Entries: []*parse.Error{
			{
				Message: "ERR1",
				Context: diag.Context{Name: "file1", Begin: 5, End: 7}},
			{
				Message: "ERR2",
				Context: diag.Context{Name: "file2", Begin: 15, End: 16}},
		}}).Rets(
			[]byte(`[{"fileName":"file1","start":5,"end":7,"message":"ERR1"},` +
				`{"fileName":"file2","start":15,"end":16,"message":"ERR2"}]`),
		),
		tt.Args(errors.New("ERR")).Rets([]byte(`[{"message":"ERR"}]`)),
	})
}
