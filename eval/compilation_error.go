package eval

import (
	"bytes"
	"fmt"

	"github.com/elves/elvish/util"
)

// CompilationError represents a compilation error and can pretty print it.
type CompilationError struct {
	Message string
	Context util.SourceRange
}

func (ce *CompilationError) Error() string {
	return fmt.Sprintf("compilation error: %d-%d in %s: %s",
		ce.Context.Begin, ce.Context.End, ce.Context.Name, ce.Message)
}

// Pprint pretty-prints a compilation error.
func (ce *CompilationError) Pprint(indent string) string {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "Compilation error: \033[31;1m%s\033[m\n", ce.Message)
	fmt.Fprint(buf, indent+"  ")
	ce.Context.Pprint(buf, indent+"    ")

	return buf.String()
}
