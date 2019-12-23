package shell

import (
	"errors"
	"testing"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/tt"
)

func TestErrorsToJSON(t *testing.T) {
	tt.Test(t, tt.Fn("errorToJSON", errorToJSON), tt.Table{
		tt.Args(eval.NewCompilationError(
			"ERR",
			diag.Context{Name: "file", Ranging: diag.Ranging{From: 5, To: 7}}),
		).Rets(
			[]byte(`[{"fileName":"file","start":5,"end":7,"message":"ERR"}]`),
		),
		tt.Args(&parse.MultiError{Entries: []*parse.Error{
			{
				Message: "ERR1",
				Context: diag.Context{Name: "file1", Ranging: diag.Ranging{From: 5, To: 7}}},
			{
				Message: "ERR2",
				Context: diag.Context{Name: "file2", Ranging: diag.Ranging{From: 15, To: 16}}},
		}}).Rets(
			[]byte(`[{"fileName":"file1","start":5,"end":7,"message":"ERR1"},` +
				`{"fileName":"file2","start":15,"end":16,"message":"ERR2"}]`),
		),
		tt.Args(errors.New("ERR")).Rets([]byte(`[{"message":"ERR"}]`)),
	})
}
