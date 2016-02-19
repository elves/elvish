package util

import (
	"bytes"
	"fmt"
	"strings"
)

// PosError is an error associated with a position range.
type PosError struct {
	Begin int
	End   int
	Err   error
}

func (pe *PosError) Error() string {
	return fmt.Sprintf("%d-%d: %s", pe.Begin, pe.End, pe.Err.Error())
}

// Pprint pretty-prints a PosError with a header indicating the source and type
// of the error, the error text and the affected line with an additional line
// that points an arrow at the affected column.
func (pe *PosError) Pprint(srcname, errtype, src string) string {
	lineno, colno, line := FindContext(src, pe.Begin)

	buf := new(bytes.Buffer)
	// Source and type of the error
	fmt.Fprintf(buf, "\033[1m%s:%d:%d: \033[31m%s:", srcname, lineno+1, colno+1, errtype)
	// Message
	fmt.Fprintf(buf, "\033[m\033[1m%s\033[m\n", pe.Err.Error())
	// Affected line
	// TODO Handle long lines
	fmt.Fprintf(buf, "%s\n", line)
	// Column indicator
	// TODO Handle multi-width characters
	fmt.Fprintf(buf, "%s\033[32;1m^\033[m\n", strings.Repeat(" ", colno))
	return buf.String()
}
