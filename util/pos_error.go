package util

import (
	"bytes"
	"fmt"
)

// PosError is an error associated with a position range.
type PosError struct {
	Begin int
	End   int
	Err   error
}

func (pe *PosError) Error() string {
	return fmt.Sprintf("%d-%d: %s", pe.Begin, pe.End, pe.msg())
}

// Pprint pretty-prints a PosError with a header indicating the source and type
// of the error, the error text and the affected line with an additional line
// that points an arrow at the affected column.
func (pe *PosError) Pprint(srcname, errtype, src string) string {
	buf := new(bytes.Buffer)
	// Error message
	fmt.Fprintf(buf, "%s: \033[31;1m%s\033[m\n", errtype, pe.msg())
	// Trace back
	//buf.WriteString("  ")
	//buf.WriteString("Traceback:\n  ")
	te := TracebackEntry{srcname, src, pe.Begin, pe.End}
	te.Pprint(buf, "  ")

	return buf.String()
}

func (pe *PosError) msg() string {
	if pe.Err != nil {
		return pe.Err.Error()
	} else {
		return "<nil>"
	}
}
