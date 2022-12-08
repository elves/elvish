package diag

import (
	"fmt"

	"src.elv.sh/pkg/strutil"
)

// Error represents an error with context that can be showed.
type Error struct {
	Type    string
	Message string
	Context Context
}

// Error returns a plain text representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s:%s: %s",
		e.Type, e.Context.Name, e.Context.culprit.describeStart(), e.Message)
}

// Range returns the range of the error.
func (e *Error) Range() Ranging {
	return e.Context.Range()
}

var (
	messageStart = "\033[31;1m"
	messageEnd   = "\033[m"
)

// Show shows the error.
func (e *Error) Show(indent string) string {
	return fmt.Sprintf("%s: %s%s%s\n%s%s",
		strutil.Title(e.Type), messageStart, e.Message, messageEnd,
		indent+"  ", e.Context.ShowCompact(indent+"  "))
}
