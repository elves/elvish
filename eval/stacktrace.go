package eval

import "github.com/elves/elvish/diag"

// stackTrace represents a stack trace as a linked list. Inner stacks appear
// first. Since pipelines can call multiple functions in parallel, all the
// stackTrace nodes form a DAG.
type stackTrace struct {
	entry *diag.Context
	next  *stackTrace
}
