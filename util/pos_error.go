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

// Pprint pretty-prints a PosError.
func (pe *PosError) Pprint(srcname, errtype, src string) string {
	buf := new(bytes.Buffer)
	// Error message
	fmt.Fprintf(buf, "%s: \033[31;1m%s\033[m\n", errtype, pe.msg())
	// Position
	te := Traceback{srcname, src, pe.Begin, pe.End, nil}
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
